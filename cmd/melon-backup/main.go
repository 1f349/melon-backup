package main

import (
	"context"
	"flag"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
	"os"
)

func main() {
	log.Info("melon-backup  (C) 1f349 2026  GPL Version 3 License")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&daemonCmd{}, "")
	subcommands.Register(&generateCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
