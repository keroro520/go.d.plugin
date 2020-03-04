package ckb

import "github.com/netdata/go-orchestrator/module"

// `Path` and `Journal` are mutually exclusive
type Config struct {
	Path    string `yaml:"path"`
	Journal string `yaml:"journal"`
	Charts  module.Charts `yaml:"charts"`
}

func defaultConfig() Config {
	return Config {
		Path: "",
		Journal: "",
		Charts: make(module.Charts, 0, 1000),
	}
}
