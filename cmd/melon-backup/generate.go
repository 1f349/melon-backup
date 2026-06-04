package main

import (
	"context"
	"flag"
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
	"os"
)

type generateCmd struct {
	configPath string
	example    bool
}

func (g *generateCmd) Name() string {
	return "generate"
}

func (g *generateCmd) Synopsis() string {
	return "Generate a config file"
}

func (g *generateCmd) Usage() string {
	return `generate [-config <config file>] [-example]
  Generate an example config file.
`
}

func (g *generateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&g.configPath, "config", "", "/path/to/config.yml : path to the configuration file")
	f.BoolVar(&g.example, "example", false, "generate example configuration file with default values")
}

func (g *generateCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	log.Info("Generating config file...")

	if g.configPath == "" {
		log.Error("Configuration file path is required")
		return subcommands.ExitUsageError
	}

	openConf, err := os.OpenFile(g.configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
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

	conf.Generate(openConf, nil, g.example)

	return subcommands.ExitSuccess
}
