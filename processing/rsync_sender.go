package processing

import (
	"github.com/1f349/melon-backup/comm"
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/proxy"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
	"io"
	"os/exec"
)

type RsyncSender struct {
	cnf     conf.ConfigYAML
	lstnr   *proxy.Listener
	cmd     *exec.Cmd
	pipeOut io.ReadCloser
	pipeErr io.ReadCloser
}

func NewRsyncSender(cnf conf.ConfigYAML, conn *comm.Client) *RsyncSender {
	lstnr, err := proxy.NewListener(conn, cnf)
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return nil
	}
	log.Info("Proxy Listening Started!")
	return &RsyncSender{cnf: cnf, lstnr: lstnr, cmd: utils.CreateCmd(cnf.RSyncCommand, "RSYNC_PASSWORD="+cnf.Security.RSyncPassword)}
}

func (s *RsyncSender) StartAndWait() {
	log.Info("RSync Started...")
	defer func() {
		s.lstnr.Close()
		log.Info("Proxy Listener Closed!")
	}()
	if s.cmd == nil {
		log.Error("No command!")
		return
	}
	var err error
	if conf.Debug {
		s.pipeErr, err = s.cmd.StderrPipe()
		if err != nil {
			log.Error(err)
			return
		}
	}
	s.pipeOut, err = s.cmd.StdoutPipe()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return
	}
	err = s.cmd.Start()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return
	}
	if conf.Debug {
		go s.readSTDErr()
	}
	go s.readSTDOut()
	err = s.cmd.Wait()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return
	}
	log.Info("RSync Finished!")
}

func (s *RsyncSender) readSTDErr() {
	defer func() {
		bts, err := io.ReadAll(s.pipeErr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Error(string(bts))
	}()
	buff := make([]byte, s.cnf.GetTarBufferSize())
	var br int
	var err error
	for s.cmd.ProcessState == nil {
		br, err = s.pipeErr.Read(buff)
		if err != nil {
			log.Error(err)
			return
		} else {
			log.Error(string(buff[:br]))
		}
	}
}

func (s *RsyncSender) readSTDOut() {
	defer func() {
		bts, err := io.ReadAll(s.pipeOut)
		if err != nil {
			if conf.Debug {
				log.Error(err)
			}
			return
		}
		log.Info(bts)
	}()
	buff := make([]byte, s.cnf.GetTarBufferSize())
	var br int
	var err error
	for s.cmd.ProcessState == nil {
		br, err = s.pipeOut.Read(buff)
		if err != nil {
			if conf.Debug {
				log.Error(err)
			}
			return
		}
		log.Info(string(buff[:br]))
	}
}
