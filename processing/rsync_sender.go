package processing

import (
	"github.com/1f349/melon-backup/comm"
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/proxy"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
	"os/exec"
)

type RsyncSender struct {
	cnf   conf.ConfigYAML
	lstnr *proxy.Listener
	cmd   *exec.Cmd
}

func NewRsyncSender(cnf conf.ConfigYAML, conn *comm.Client, debug bool) *RsyncSender {
	lstnr, err := proxy.NewListener(conn, cnf, debug)
	if err != nil {
		if debug {
			log.Error(err)
		}
		return nil
	}
	log.Info("Proxy Listening Started!")
	return &RsyncSender{cnf: cnf, lstnr: lstnr, cmd: utils.CreateCmd(cnf.RSyncCommand, "RSYNC_PASSWORD="+cnf.Security.RSyncPassword)}
}

func (s *RsyncSender) StartAndWait(debug bool) {
	log.Info("RSync Started...")
	defer func() {
		s.lstnr.Close()
		log.Info("Proxy Listener Closed!")
	}()
	if s.cmd == nil {
		log.Error("No command!")
		return
	}
	bts, err := s.cmd.CombinedOutput()
	if err != nil {
		if debug {
			log.Error(err)
		}
	}
	log.Info(string(bts))
	log.Info("RSync Finished!")
}
