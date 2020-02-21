package ckb

import (
	"github.com/netdata/go-orchestrator/module"
	"io"
	"os"
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
	c.Infof("using metrics log %s", c.Path)
	return true
}

func (c *Ckb) Check() bool {
	// Note: these inits are here to make auto detection retry working
	c.Cleanup()
	file, err := os.Open(c.Path)
	file.Seek(0, io.SeekEnd)
	if err != nil {
		c.Errorf("error on opening log file: %v", err)
		return false
	}

	c.parser = NewParser(file)
	return true
}

func (Ckb) Charts() *Charts { return charts.Copy() }

func (c *Ckb) Collect() map[string]int64 {
	for dimId := range c.metrics {
		c.metrics[dimId] = 0
	}

	for {
		metric, err := c.parser.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			c.Errorf("error on reading logfile: %v", err)
			break
		} else {
			if _, ok := c.metrics[metric.Topic]; ok {
				c.Debugf("read a metric: %v", metric)
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
