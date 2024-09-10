package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
	"slices"
)

func StartServices(cnf conf.ConfigYAML, previouslyStopped []string, sent []string, debug bool) {
	if len(cnf.Services.ReloadCommand) > 0 {
		log.Info("Reloading Services...")
		cmdRl := utils.CreateCmd(cnf.Services.ReloadCommand)
		err := cmdRl.Run()
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
	}
	if len(cnf.Services.StartCommand) > 0 {
		if cnf.Services.Restore {
			toRestore := slices.DeleteFunc(sent, func(s string) bool {
				return !slices.Contains(previouslyStopped, s)
			})
			log.Info("Service Restore Task Started...")
			for _, n := range toRestore {
				log.Info("Starting: " + n)
				cmd := utils.CreateCmd(append(cnf.Services.StartCommand, n))
				err := cmd.Run()
				if err != nil {
					log.Info("Failed to start: " + n)
					if debug {
						log.Error(err)
					}
				}
			}
			log.Info("Service Restore Task Completed!")
		}
		if cnf.Services.StartNew {
			log.Info("Service Start New Task Started...")
			toStartNew := slices.DeleteFunc(sent, func(s string) bool {
				return slices.Contains(previouslyStopped, s) || slices.Contains(cnf.Services.List, s)
			})
			for _, n := range toStartNew {
				log.Info("Starting: " + n)
				cmd := utils.CreateCmd(append(cnf.Services.StartCommand, n))
				err := cmd.Run()
				if err != nil {
					log.Info("Failed to start: " + n)
					if debug {
						log.Error(err)
					}
				}
			}
			log.Info("Service Start New Task Completed!")
		}
	}
}
