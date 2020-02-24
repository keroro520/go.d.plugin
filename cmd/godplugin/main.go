package main

import (
	"fmt"
	"os"
	"path"

	"github.com/jessevdk/go-flags"
	"github.com/netdata/go-orchestrator"
	"github.com/netdata/go-orchestrator/cli"
	"github.com/netdata/go-orchestrator/logger"
	"github.com/netdata/go-orchestrator/pkg/multipath"

	_ "github.com/netdata/go.d.plugin/modules/ckb"
)

var (
	cd, _       = os.Getwd()
	configPaths = multipath.New(
		os.Getenv("NETDATA_USER_CONFIG_DIR"),
		os.Getenv("NETDATA_STOCK_CONFIG_DIR"),
		path.Join(cd, "/../../../../etc/netdata"),
		path.Join(cd, "/../../../../usr/lib/netdata/conf.d"),
	)
)

var version = "unknown"

func main() {
	opt := parseCLI()

	if opt.Version {
		fmt.Println(fmt.Sprintf("go.d.plugin, version: %s", version))
		return
	}
	if opt.Debug {
		logger.SetSeverity(logger.DEBUG)
	}

	plugin := newPlugin(opt)

	if !plugin.Setup() {
		os.Exit(1)
	}

	plugin.Serve()
}

func newPlugin(opt *cli.Option) *orchestrator.Orchestrator {
	plugin := orchestrator.New()
	plugin.Name = "go.d"
	plugin.Option = opt
	plugin.ConfigPath = configPaths
	return plugin
}

func parseCLI() *cli.Option {
	opt, err := cli.Parse(os.Args)
	if err != nil {
		if isHelp(err) {
			os.Exit(0)
		}
		os.Exit(1)
	}
	return opt
}

func isHelp(err error) bool {
	flagsErr, ok := err.(*flags.Error)
	return ok && flagsErr.Type == flags.ErrHelp
}
