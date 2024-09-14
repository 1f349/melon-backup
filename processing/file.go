package processing

import (
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"io"
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

func NewFileTask(conn net.Conn, cnf conf.ConfigYAML) *File {
	flp, err := os.OpenFile(cnf.GetStoreFile(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		if conf.Debug {
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
	log.Info("Started Writing to file: " + cnf.GetStoreFile())
	go fl.writeFileOut()
	if cnf.Net.KeepAliveTime > time.Millisecond {
		go fl.sendKeepAlives()
	}
	return fl
}

func (f *File) WaitOnCompletion() {
	<-f.endChan
	log.Info("Finished Writing to file: " + f.cnf.GetStoreFile())
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

func (f *File) writeFileOut() {
	defer func() {
		_ = f.file.Close()
		_ = f.conn.Close()
		close(f.endChan)
	}()
	buff := make([]byte, f.cnf.GetTarBufferSize())
	_, err := io.CopyBuffer(f.file, f.conn, buff)
	if err != nil && conf.Debug {
		log.Error(err)
	}
}
