package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
	"io"
	"net"
	"os/exec"
)

type Tar struct {
	cmd     *exec.Cmd
	cnf     conf.ConfigYAML
	conn    net.Conn
	pipeOut io.ReadCloser
	pipeErr io.ReadCloser
}

func NewTarTask(conn net.Conn, cnf conf.ConfigYAML, debug bool) *Tar {
	cmd := utils.CreateCmd(cnf.TarCommand)
	if cmd == nil {
		log.Error("No command!")
		return nil
	}
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
	err = cmd.Start()
	if err != nil {
		if debug {
			log.Error(err)
		}
		return nil
	}
	tar := &Tar{
		cmd:     cmd,
		cnf:     cnf,
		conn:    conn,
		pipeOut: stdout,
		pipeErr: stderr,
	}
	log.Info("Tar Operation Started!")
	go tar.readSTDErr()
	go tar.readSTDOut(debug)
	go tar.eatAllKeepAlives()
	return tar
}

func (t *Tar) WaitOnCompletion(debug bool) {
	err := t.cmd.Wait()
	if err != nil {
		if debug {
			log.Error(err)
		}
		return
	}
	log.Info("Tar Operation Completed!")
}

func (t *Tar) eatAllKeepAlives() {
	buff := make([]byte, t.cnf.GetTarBufferSize())
	var err error
	for t.cmd.ProcessState == nil {
		_, err = t.conn.Read(buff)
		if err != nil {
			return
		}
	}
}

func (t *Tar) readSTDErr() {
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

func (t *Tar) readSTDOut(debug bool) {
	defer func() {
		_, err := io.ReadAll(t.pipeOut)
		if err != nil && debug {
			log.Error(err)
		}
		if t.cmd.ProcessState == nil {
			err = t.cmd.Process.Kill()
			if err != nil && debug {
				log.Error(err)
			}
		}
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
		} else {
			_, err := t.conn.Write(buff[:br])
			if err != nil {
				if debug {
					log.Error(err)
				}
				return
			}
		}
	}
}
