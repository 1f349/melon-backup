package processing

import (
	"github.com/1f349/melon-backup/comm"
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/proxy"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
)

type RsyncIngester struct {
	cnf   conf.ConfigYAML
	lstnr *proxy.Multiplexer
}

func NewRsyncIngester(cnf conf.ConfigYAML, conn *comm.Client, debug bool) *RsyncIngester {
	if cnf.Services.ManageRSync {
		cmd := utils.CreateCmd(append(cnf.Services.StartCommand, "rsync.service"))
		err := cmd.Run()
		if err != nil {
			if debug {
				log.Error(err)
			}
			return nil
		}
	}
	return &RsyncIngester{cnf: cnf, lstnr: proxy.NewMultiplexer(conn, cnf, debug)}
}

func (r *RsyncIngester) Wait(debug bool) {
	<-r.lstnr.GetCloseWaiter()
	if r.cnf.Services.ManageRSync {
		cmd := utils.CreateCmd(append(r.cnf.Services.StopCommand, "rsync.service"))
		err := cmd.Run()
		if err != nil {
			if debug {
				log.Error(err)
			}
		}
	}
}
