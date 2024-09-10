package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"net"
	"os"
)

type UnFile struct {
	cnf     conf.ConfigYAML
	conn    net.Conn
	file    *os.File
	endChan chan struct{}
}

func NewUnFileTask(conn net.Conn, cnf conf.ConfigYAML, debug bool) *UnFile {
	flp, err := os.Open(cnf.GetStoreFile())
	if err != nil {
		if debug {
			log.Error(err)
		}
		return nil
	}
	fl := &UnFile{
		cnf:     cnf,
		conn:    conn,
		file:    flp,
		endChan: make(chan struct{}),
	}
	go fl.readFileIn(debug)
	go fl.eatAllKeepAlives()
	return fl
}

func (f *UnFile) WaitOnCompletion() {
	<-f.endChan
}

func (t *UnFile) eatAllKeepAlives() {
	buff := make([]byte, t.cnf.GetTarBufferSize())
	var err error
	for {
		_, err = t.conn.Read(buff)
		if err != nil {
			return
		}
	}
}

func (f *UnFile) readFileIn(debug bool) {
	defer func() {
		_ = f.file.Close()
		_ = f.conn.Close()
		close(f.endChan)
	}()
	buff := make([]byte, f.cnf.GetTarBufferSize())
	var br int
	var err error
	for {
		br, err = f.file.Read(buff)
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
		_, err = f.conn.Write(buff[:br])
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
	}
}
