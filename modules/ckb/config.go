package ckb

import "github.com/netdata/go-orchestrator/module"

// `Path` and `Journal` are mutually exclusive
type Config struct {
	LogToFile string `yaml:"log_to_file"`
	LogToJournal string `yaml:"log_to_journal"`
	Charts  module.Charts `yaml:"charts"`
}

func defaultConfig() Config {
	return Config {
		LogToFile: "",
		LogToJournal: "",
		Charts: make(module.Charts, 0, 1000),
	}
}
