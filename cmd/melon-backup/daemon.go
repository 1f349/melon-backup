package main

import (
	"context"
	"flag"
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/processing"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
	"gopkg.in/yaml.v3"
	"os"
)

type daemonCmd struct {
	configPath  string
	debug       bool
	multiAccess bool
}

func (d *daemonCmd) Name() string {
	return "daemon"
}

func (d *daemonCmd) Synopsis() string {
	return "Run the daemon"
}

func (d *daemonCmd) Usage() string {
	return `daemon [-config <config file>] [-debug] [-multi]
  Run the daemon using the specified config file.
`
}

func (d *daemonCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&d.configPath, "config", "", "/path/to/config.yml : path to the configuration file")
	f.BoolVar(&d.debug, "debug", false, "enable debug mode")
	f.BoolVar(&d.multiAccess, "multi", false, "allow acceptance of clients when listening till the first valid connection")
}

func (d *daemonCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	log.Info("Starting daemon ...")

	conf.Debug = d.debug

	if d.configPath == "" {
		log.Error("Configuration file path is required")
		return subcommands.ExitUsageError
	}

	openConf, err := os.Open(d.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("Missing config file")
		} else {
			log.Error("Open config file: ", err)
		}
		return subcommands.ExitFailure
	}

	var cnf conf.ConfigYAML
	err = yaml.NewDecoder(openConf).Decode(&cnf)
	_ = openConf.Close()
	if err != nil {
		log.Error("Invalid config file: ", err)
		return subcommands.ExitFailure
	}

	rv := processing.Start(cnf, d.multiAccess)

	if rv != 0 {
		return subcommands.ExitStatus(rv + 10)
	}
	return subcommands.ExitStatus(rv)
}
