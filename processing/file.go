package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"net"
	"os"
	"time"
)

type File struct {
	cnf     conf.ConfigYAML
	conn    net.Conn
	file    *os.File
	endChan chan struct{}
}

func NewFileTask(conn net.Conn, cnf conf.ConfigYAML, debug bool) *File {
	flp, err := os.OpenFile(cnf.GetStoreFile(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		if debug {
			log.Error(err)
		}
		return nil
	}
	fl := &File{
		cnf:     cnf,
		conn:    conn,
		file:    flp,
		endChan: make(chan struct{}),
	}
	go fl.writeFileOut(debug)
	if cnf.Net.KeepAliveTime > time.Millisecond {
		go fl.sendKeepAlives()
	}
	return fl
}

func (f *File) WaitOnCompletion() {
	<-f.endChan
}

func (t *File) sendKeepAlives() {
	kAlive := time.NewTimer(t.cnf.Net.KeepAliveTime)
	defer func() {
		kAlive.Stop()
		_ = t.conn.Close()
	}()
	var err error
	for {
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

func (f *File) writeFileOut(debug bool) {
	defer func() {
		_ = f.file.Close()
		_ = f.conn.Close()
		close(f.endChan)
	}()
	buff := make([]byte, f.cnf.GetTarBufferSize())
	var br int
	var err error
	for {
		br, err = f.conn.Read(buff)
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
		_, err = f.file.Write(buff[:br])
		if err != nil {
			if debug {
				log.Error(err)
			}
			return
		}
	}
}
