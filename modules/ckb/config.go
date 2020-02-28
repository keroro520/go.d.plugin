package ckb

// `Path` and `Journal` are mutually exclusive
type Config struct {
	Path    string `yaml:"path"`
	Journal string `yaml:"journal"`
}

func defaultConfig() Config {
	return Config {
		Path: "",
		Journal: "",
	}
}
