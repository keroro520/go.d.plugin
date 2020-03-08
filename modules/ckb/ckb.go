package ckb

import (
    "github.com/coreos/go-systemd/sdjournal"
	"github.com/netdata/go-orchestrator/module"
	"io"
	"os"
	"time"
	"strings"
	"fmt"
	"regexp"
)

type DimIdAlgo struct {
	ID string
	Algo string
}

func init() {
	module.Register("ckb", module.Creator{
		Defaults: module.Defaults{Disabled: true},
		Create:   func() module.Module { return New() },
	})
}

func New() *Ckb {
	return &Ckb{
		Config:           defaultConfig(),
		metrics:          make(map[string]int64),
		charts: nil,
		FieldToDimIdAlgo: make(map[string][]DimIdAlgo),
	}
}

type Ckb struct {
	module.Base
	Config `yaml:",inline"`

	parser           *Parser
	metrics          map[string]int64
	charts           module.Charts
	FieldToDimIdAlgo map[string][]DimIdAlgo
}

func (c *Ckb) Init() bool {
	c.charts = *c.Charts()
	
	for _, chart := range c.charts {
		for _, dim := range chart.Dims {
			c.metrics[dim.ID] = 0
		}
	}

	return true
}

func (c *Ckb) Check() bool {
	// Note: these inits are here to make auto detection retry working
	c.Cleanup()

	if c.LogToJournal != "" {
		if !strings.HasSuffix(c.LogToJournal, ".service") {
			c.LogToJournal = c.LogToJournal+ ".service"
		}
		config := sdjournal.JournalReaderConfig{
			Since: time.Second,
			Matches: []sdjournal.Match{
				{
					Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
					Value: c.LogToJournal,
				},
			},
		}
		journal, err := sdjournal.NewJournalReader(config)
		if err != nil {
			c.Errorf("error on creating journal reader: %v", err)
			return false
		}
		c.Infof("using journal log like `sudo journalctl -u %s -f`", c.LogToJournal)
		c.parser = NewJournalParser(journal)
	} else {
		file, err := os.Open(c.LogToFile)
		if err != nil {
			c.Errorf("error on opening log file: %v", err)
			return false
		}
		file.Seek(0, io.SeekEnd)
		c.Infof("using file log %s", c.LogToFile)
		c.parser = NewFileParser(file)
	}

	return true
}

func (c *Ckb) Charts() *module.Charts {
	if c.charts != nil {
		return &c.charts
	}

	charts := c.Config.Charts

	// Use chart.title as chart.id if empty
	for _, chart := range charts {
		if chart.ID == "" {
			chart.ID = chart.Title
		}
	}

	// Parse dim.name into below formats:
	//   - "<name>:last(<field>)", last absolute number during checking period
	//   - "<name>:inc(<field>)", incremental to the latest value at the last one round checking
	//   - "<name>:max(<field>)", max during checking period, and will be reset to zero every time
	//   - "<name>:min(<field>)", min during checking period, and will be reset to zero every time
	//   - "<name>:sum(<field>)", sum during checking period, and will be reset to zero every time
	//   - "<name>:total(<field>)", sum all the history
	pattern := regexp.MustCompile(`^([a-zA-Z0-9_.]*):(last|inc|max|min|sum|total)\(([a-zA-Z0-9_.]*)\)$`)
	for _, chart := range charts {
		for _, dim := range chart.Dims {
			matches := pattern.FindStringSubmatch(dim.Name)
			if matches == nil || len(matches) != 4 {
				c.Errorf("Invalid dim.Name: %s, ignore it", dim.Name)
				continue
			}
			name, algo, field := matches[1], matches[2], matches[3]

			dim.Name = name
			if dim.ID == "" {
				dim.ID = fmt.Sprintf("%s.%s", chart.Title, dim.Name)
			}

			switch algo {
			case "inc":
				dim.Algo = module.Incremental
			default:
				dim.Algo = module.Absolute
			}

			dimIdAlgo := DimIdAlgo{ID: dim.ID, Algo: algo}
			if l, ok := c.FieldToDimIdAlgo[field]; ok {
				c.FieldToDimIdAlgo[field] = append(l, dimIdAlgo)
			} else {
				l := make([]DimIdAlgo, 0)
				c.FieldToDimIdAlgo[field] = append(l, dimIdAlgo)
			}

			c.Infof("[ckb_config] New Dim: chart.id(%s), dim.id(%s), dim.name(%s), dim.measurement(%s), dim.algo(%s), dim.handler(%s)", chart.ID, dim.ID, dim.Name, field, dim.Algo, dimIdAlgo.Algo)
		}
	}

	return charts.Copy()
}

func (c *Ckb) Collect() map[string]int64 {
	if c.parser == nil {
		return c.metrics
	}

	// Reset metrics for reset-to-zero algorithms
	for _, dimIdAlgos := range c.FieldToDimIdAlgo {
		for _, dimIdAlgo := range dimIdAlgos {
			switch dimIdAlgo.Algo {
			case "max","min","sum":
				c.metrics[dimIdAlgo.ID] = 0
			}
		}
	}

	for {
		metric, err := c.parser.ReadLine()
		if err == io.EOF {        // EOF
			break
		} else if metric == nil { // Unmatched or parse error
			continue
		} else if err == nil {    // A metric entry
			c.preprocess(metric)

			for field, value := range metric.Fields {
				if dimIdAlgos, ok := c.FieldToDimIdAlgo[field]; ok {
					for _, dimIdAlgo := range dimIdAlgos {
						value := int64(value)
						dimId, algo := dimIdAlgo.ID, dimIdAlgo.Algo
						switch algo {
						case "inc","last":
							c.metrics[dimId] = value
						case "min":
							if old, ok := c.metrics[dimId]; ok && (old == 0 || old > value) {
								c.metrics[dimId] = value
							}
						case "max":
							if old, ok := c.metrics[dimId]; ok && (old == 0 || old < value) {
								c.metrics[dimId] = value
							}
						case "sum","total":
							c.metrics[dimId] += value
						}
					}
				}
			}
		}
	}

	return c.metrics
}

func (c *Ckb) preprocess(metric *Metric) {
	if len(metric.Fields) == 0 {
		metric.Fields[metric.Topic] = 1
	}

	newFields := make(map[string]uint64, 0)
	for field, value := range metric.Fields {
		newFields[fmt.Sprintf("%s.%s", metric.Topic, field)] = value
	}
	metric.Fields = newFields
}

func (c *Ckb) Cleanup() {
	if c.parser != nil {
		c.parser.Close()
	}
}
