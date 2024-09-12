package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"io"
	"os/exec"
	"time"
)

type ConnToCommand struct {
	cmd         *exec.Cmd
	commandName string
	cnf         conf.ConfigYAML
	conn        io.ReadWriteCloser
	pipeIn      io.WriteCloser
	pipeOut     io.ReadCloser
	pipeErr     io.ReadCloser
	endChan     chan struct{}
}

func NewConnToCommandTask(conn io.ReadWriteCloser, keepAlive bool, name string, cmd *exec.Cmd, cnf conf.ConfigYAML) *ConnToCommand {
	if cmd == nil {
		log.Error("No command!")
		return nil
	}
	var err error
	var stderr io.ReadCloser
	if conf.Debug {
		stderr, err = cmd.StderrPipe()
		if err != nil {
			log.Error(err)
			return nil
		}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return nil
	}
	stdin, err := cmd.StdinPipe()
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
	commToCmd := &ConnToCommand{
		cmd:         cmd,
		commandName: name,
		cnf:         cnf,
		conn:        conn,
		pipeOut:     stdout,
		pipeErr:     stderr,
		pipeIn:      stdin,
		endChan:     make(chan struct{}),
	}
	log.Info(name + " Operation Started!")
	if conf.Debug {
		go commToCmd.readSTDErr()
	}
	go commToCmd.writeSTDIn()
	go commToCmd.readSTDOut()
	if keepAlive && cnf.Net.KeepAliveTime > time.Millisecond {
		go commToCmd.sendKeepAlives()
	}
	return commToCmd
}

func (t *ConnToCommand) WaitOnCompletion() {
	<-t.endChan
	err := t.cmd.Wait()
	if err != nil {
		if conf.Debug {
			log.Error(err)
		}
		return
	}
	log.Info(t.commandName + " Operation Completed!")
}

func (t *ConnToCommand) sendKeepAlives() {
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

func (t ConnToCommand) writeSTDIn() {
	defer func() {
		_ = t.pipeIn.Close()
		_ = t.conn.Close()
		close(t.endChan)
	}()
	buff := make([]byte, t.cnf.GetTarBufferSize())
	var br int
	var err error
	for t.cmd.ProcessState == nil {
		br, err = t.conn.Read(buff)
		if err != nil {
			if conf.Debug {
				log.Error(err)
			}
			return
		}
		_, err = t.pipeIn.Write(buff[:br])
		if err != nil {
			if conf.Debug {
				log.Error(err)
			}
			return
		}
	}
}

func (t *ConnToCommand) readSTDErr() {
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

func (t *ConnToCommand) readSTDOut() {
	defer func() {
		bts, err := io.ReadAll(t.pipeOut)
		if err != nil {
			if conf.Debug {
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
			if conf.Debug {
				log.Error(err)
			}
			return
		}
		log.Info(string(buff[:br]))
	}
}
