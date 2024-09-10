package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
)

func StopServices(cnf conf.ConfigYAML, debug bool) []string {
	if cnf.Services.Stop {
		toRet := make([]string, 0, len(cnf.Services.List))
		log.Info("Service Stop Task Started...")
		for _, n := range cnf.Services.List {
			log.Info("Stopping: " + n)
			cmd := utils.CreateCmd(append(cnf.Services.StopCommand, n))
			err := cmd.Run()
			if err != nil {
				log.Info("Failed to stop: " + n)
				if debug {
					log.Error(err)
				}
			} else {
				toRet = append(toRet, n)
			}
		}
		log.Info("Service Stop Task Completed!")
		return toRet
	}
	return nil
}
