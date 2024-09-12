package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"os/exec"
)

type Command struct {
	cnf         conf.ConfigYAML
	cmd         *exec.Cmd
	commandName string
}

func NewCommandTask(cnf conf.ConfigYAML, cmd *exec.Cmd, name string) *Command {
	if cmd == nil {
		log.Error("No command!")
		return nil
	}
	return &Command{cnf: cnf, cmd: cmd, commandName: name}
}

func (c *Command) StartAndWait() {
	log.Info(c.commandName + " Started...")
	if c.cmd == nil {
		log.Error("No command!")
		return
	}
	bts, err := c.cmd.CombinedOutput()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
	}
	log.Info(string(bts))
	log.Info(c.commandName + " Finished!")
}
