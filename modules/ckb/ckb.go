package ckb

import (
    "github.com/coreos/go-systemd/sdjournal"
	"github.com/netdata/go-orchestrator/module"
	"io"
	"os"
	"time"
	"strings"
)

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
}

// `Init` initializes metric-items and charts
// `Check` initializes file

func (c *Ckb) Init() bool {
	for _, chart := range charts {
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

func (Ckb) Charts() *Charts { return charts.Copy() }

func (c *Ckb) Collect() map[string]int64 {
	if c.parser == nil {
		return c.metrics
	}

	for dimId := range c.metrics {
		c.metrics[dimId] = 0
	}

	for {
		metric, err := c.parser.ReadLine()
		if err == io.EOF {        // EOF
			break
		} else if metric == nil { // Unmatched or parse error
			continue
		} else if err == nil {    // Matched
			if _, ok := c.metrics[metric.Topic]; ok {
				c.metrics[metric.Topic] += 1
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
