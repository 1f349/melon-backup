package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
)

func StopServices(cnf conf.ConfigYAML) []string {
	if cnf.GetMode() != conf.Store && cnf.GetMode() != conf.UnStore && cnf.Services.Stop && len(cnf.Services.StopCommand) > 0 {
		toRet := make([]string, 0, len(cnf.Services.List))
		log.Info("Service Stop Task Started...")
		for z := len(cnf.Services.List) - 1; z >= 0; z-- {
			n := cnf.Services.List[z]
			log.Info("Stopping: " + n)
			cmd := utils.CreateCmd(append(cnf.Services.StopCommand, n))
			err := cmd.Run()
			if err != nil {
				log.Info("Failed to stop: " + n)
				if conf.Debug {
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
