package main

import (
	"context"
	"flag"
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
	"os"
)

type editCmd struct {
	configPath string
}

func (g *editCmd) Name() string {
	return "edit"
}

func (g *editCmd) Synopsis() string {
	return "Edit a config file"
}

func (g *editCmd) Usage() string {
	return `edit [-config <config file>]
  Edit a config file.
`
}

func (g *editCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&g.configPath, "config", "", "/path/to/config.yml : path to the configuration file")
}

func (g *editCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	log.Info("Reading config file...")

	if g.configPath == "" {
		log.Error("Configuration file path is required")
		return subcommands.ExitUsageError
	}

	openConf, err := os.OpenFile(g.configPath, os.O_RDWR, 0640)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("Missing config file")
		} else {
			log.Error("Open config file: ", err)
		}
		return subcommands.ExitFailure
	}
	defer func() {
		_ = openConf.Close()
	}()

	sz := conf.Generate(openConf, openConf, false)

	err = openConf.Truncate(sz)
	if err != nil {
		log.Error("Truncate config file: ", err)
	}

	return subcommands.ExitSuccess
}
