package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
	"io"
	"net"
	"os/exec"
	"time"
)

type UnTar struct {
	cmd     *exec.Cmd
	cnf     conf.ConfigYAML
	conn    net.Conn
	pipeIn  io.WriteCloser
	pipeOut io.ReadCloser
	pipeErr io.ReadCloser
	endChan chan struct{}
}

func NewUnTarTask(conn net.Conn, cnf conf.ConfigYAML, debug bool) *UnTar {
	cmd := utils.CreateCmd(cnf.UnTarCommand)
	var err error
	var stderr io.ReadCloser
	if debug {
		stderr, err = cmd.StderrPipe()
		if err != nil {
			log.Error(err)
			return nil
		}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if debug {
			log.Error(err)
		}
		return nil
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		if debug {
			log.Error(err)
		}
		return nil
	}
	err = cmd.Start()
	if err != nil {
		if debug {
			log.Error(err)
		}
		return nil
	}
	tar := &UnTar{
		cmd:     cmd,
		cnf:     cnf,
		conn:    conn,
		pipeOut: stdout,
		pipeErr: stderr,
		pipeIn:  stdin,
		endChan: make(chan struct{}),
	}
	log.Info("UnTar Operation Started!")
	go tar.readSTDErr()
	go tar.writeSTDIn(debug)
	go tar.readSTDOut(debug)
	if cnf.Net.KeepAliveTime > time.Millisecond {
		go tar.sendKeepAlives()
	}
	return tar
}

func (t *UnTar) WaitOnCompletion(debug bool) {
	<-t.endChan
	err := t.cmd.Wait()
	if err != nil {
		if debug {
			log.Error(err)
		}
		return
	}
	log.Info("UnTar Operation Completed!")
}

func (t *UnTar) sendKeepAlives() {
	kAlive := time.NewTimer(t.cnf.Net.KeepAliveTime)
	defer func() {
		kAlive.Stop()
		_ = t.conn.Close()
	}()
	var err error
	for t.cmd.ProcessState == nil {
		_, err = t.conn.Write([]byte{0})
		if err != nil {
			return
		}
		select {
		case <-t.endChan:
			return
		case <-kAlive.C:
			kAlive.Reset(t.cnf.Net.KeepAliveTime)
		}
	}
}

func (t UnTar) writeSTDIn(debug bool) {
	defer func() {
		_ = t.pipeIn.Close()
		_ = t.conn.Close()
		close(t.endChan)
		if t.cmd.ProcessState == nil {
			err := t.cmd.Process.Kill()
			if err != nil && debug {
				log.Error(err)
			}
		}
	}()
	buff := make([]byte, t.cnf.GetTarBufferSize())
	var br int
	var err error
	for t.cmd.ProcessState == nil {
		br, err = t.conn.Read(buff)
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
		_, err = t.pipeIn.Write(buff[:br])
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
	}
}

func (t *UnTar) readSTDErr() {
	defer func() {
		bts, err := io.ReadAll(t.pipeErr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Error(string(bts))
	}()
	buff := make([]byte, t.cnf.GetTarBufferSize())
	var br int
	var err error
	for t.cmd.ProcessState == nil {
		br, err = t.pipeErr.Read(buff)
		if err != nil {
			log.Error(err)
			return
		} else {
			log.Error(string(buff[:br]))
		}
	}
}

func (t *UnTar) readSTDOut(debug bool) {
	defer func() {
		bts, err := io.ReadAll(t.pipeOut)
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
		log.Info(bts)
	}()
	buff := make([]byte, t.cnf.GetTarBufferSize())
	var br int
	var err error
	for t.cmd.ProcessState == nil {
		br, err = t.pipeOut.Read(buff)
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
		log.Info(string(buff[:br]))
	}
}
