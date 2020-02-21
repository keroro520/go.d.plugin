package ckb

const defaultPath = "/var/log/ckb/data/logs/run.log"

type Config struct {
	Path string `yaml:"path"`
}

func defaultConfig() Config {
	return Config{
		Path: defaultPath,
	}
}
