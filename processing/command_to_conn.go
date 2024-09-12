package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"io"
	"os/exec"
)

type CommandToConn struct {
	cmd         *exec.Cmd
	commandName string
	cnf         conf.ConfigYAML
	conn        io.ReadWriteCloser
	pipeOut     io.ReadCloser
	pipeErr     io.ReadCloser
}

func NewCommandToConnTask(conn io.ReadWriteCloser, keepAlive bool, name string, cmd *exec.Cmd, cnf conf.ConfigYAML) *CommandToConn {
	if cmd == nil {
		log.Error("No command!")
		return nil
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Error(err)
		return nil
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return nil
	}
	err = cmd.Start()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return nil
	}
	cmdToConn := &CommandToConn{
		cmd:         cmd,
		commandName: name,
		cnf:         cnf,
		conn:        conn,
		pipeOut:     stdout,
		pipeErr:     stderr,
	}
	log.Info(name + " Operation Started!")
	go cmdToConn.readSTDErr()
	go cmdToConn.readSTDOut()
	if keepAlive {
		go cmdToConn.eatAllKeepAlives()
	}
	return cmdToConn
}

func (t *CommandToConn) WaitOnCompletion() {
	err := t.cmd.Wait()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return
	}
	log.Info(t.commandName + " Operation Completed!")
}

func (t *CommandToConn) eatAllKeepAlives() {
	buff := make([]byte, t.cnf.GetTarBufferSize())
	var err error
	for t.cmd.ProcessState == nil {
		_, err = t.conn.Read(buff)
		if err != nil {
			return
		}
	}
}

func (t *CommandToConn) readSTDErr() {
	defer func() {
		bts, err := io.ReadAll(t.pipeErr)
		if err != nil {
			if conf.Debug {
				log.Error(err)
			}
			return
		}
		log.Info(string(bts))
	}()
	buff := make([]byte, t.cnf.GetTarBufferSize())
	var br int
	var err error
	for t.cmd.ProcessState == nil {
		br, err = t.pipeErr.Read(buff)
		if err != nil {
			if conf.Debug {
				log.Error(err)
			}
			return
		} else {
			log.Info(string(buff[:br]))
		}
	}
}

func (t *CommandToConn) readSTDOut() {
	defer func() {
		_, err := io.ReadAll(t.pipeOut)
		if err != nil && conf.Debug {
			log.Error(err)
		}
		if t.cmd.ProcessState == nil {
			err = t.cmd.Process.Kill()
			if err != nil && conf.Debug {
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
			if conf.Debug {
				log.Error(err)
			}
			return
		} else {
			_, err := t.conn.Write(buff[:br])
			if err != nil {
				if conf.Debug {
					log.Error(err)
				}
				return
			}
		}
	}
}
