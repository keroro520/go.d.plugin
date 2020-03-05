package ckb

import (
    "github.com/coreos/go-systemd/sdjournal"
	"github.com/netdata/go-orchestrator/module"
	"io"
	"os"
	"time"
	"strings"
	"fmt"
)

// TODO dimAlgo: Inc, Maximum and so on

func init() {
	module.Register("ckb", module.Creator{
		Defaults: module.Defaults{Disabled: true},
		Create:   func() module.Module { return New() },
	})
}

func New() *Ckb {
	return &Ckb{
		Config:  defaultConfig(),
		metrics: make(map[string]int64),
	}
}

type Ckb struct {
	module.Base
	Config `yaml:",inline"`

	parser  *Parser
	metrics map[string]int64
	charts  module.Charts
}

// `Init` initializes metric-items and charts
// `Check` initializes file

func (c *Ckb) Init() bool {
	for _, chart := range *c.Charts() {
		for _, dim := range chart.Dims {
			c.metrics[dim.ID] = 0
		}
	}
	return true
}

func (c *Ckb) Check() bool {
	// Note: these inits are here to make auto detection retry working
	c.Cleanup()

	if c.Journal != "" {
		if !strings.HasSuffix(c.Journal, ".service") {
			c.Journal = c.Journal + ".service"
		}
		config := sdjournal.JournalReaderConfig{
			Since: time.Second,
			Matches: []sdjournal.Match{
				{
					Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
					Value: c.Journal,
				},
			},
		}
		journal, err := sdjournal.NewJournalReader(config)
		if err != nil {
			c.Errorf("error on creating journal reader: %v", err)
			return false
		}
		c.Infof("using journal log like `sudo journalctl -u %s -f`", c.Journal)
		c.parser = NewJournalParser(journal)
	} else {
		file, err := os.Open(c.Path)
		if err != nil {
			c.Errorf("error on opening log file: %v", err)
			return false
		}
		file.Seek(0, io.SeekEnd)
		c.Infof("using file log %s", c.Path)
		c.parser = NewFileParser(file)
	}

	return true
}

func (c *Ckb) Charts() *module.Charts {
	charts := append(c.Config.Charts, BuiltinCharts...)
	for _, chart := range charts{
		if chart.ID == "" {
			chart.ID = chart.Title
		}

		for _, dim := range chart.Dims {
			if dim.ID == "" {
				dim.ID = fmt.Sprintf("%s-%s", chart.Title, dim.Name)
			}
		}

		if len(chart.Dims) == 0 {
			if err := chart.AddDim(&module.Dim{
				ID: fmt.Sprintf("%s-count", chart.Title),
				Name: "count",
				Algo: module.Absolute,
			}); err != nil {
				c.Errorf("Fail on creating chart %s", chart.Title)
				return nil
			}
		}
	}
	return charts.Copy()
}

func (c *Ckb) Collect() map[string]int64 {
	if c.parser == nil {
		return c.metrics
	}

	for dimId := range c.metrics {
		if strings.HasSuffix(dimId, "-count") {
			c.metrics[dimId] = 0
		}
	}

	for {
		metric, err := c.parser.ReadLine()
		if err == io.EOF {        // EOF
			break
		} else if metric == nil { // Unmatched or parse error
			continue
		} else if err == nil {    // A metric entry
			if len(metric.Fields) == 0 {	// A metric entry without any fields
				measurement := fmt.Sprintf("%s-count", metric.Topic)
				if _, ok := c.metrics[measurement]; ok {
					c.metrics[measurement] += 1
				}
			} else {
				for field, value := range metric.Fields {
					measurement := fmt.Sprintf("%s-%s", metric.Topic, field)
					if _, ok := c.metrics[measurement]; ok {
						c.metrics[measurement] = int64(value)
					}
				}
			}
		}
	}

	return c.metrics
}

func (c *Ckb) Cleanup() {
	if c.parser != nil {
		c.parser.Close()
	}
}
