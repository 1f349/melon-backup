package comm

import (
	"crypto/tls"
	"errors"
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"net"
	"strconv"
)

type Client struct {
	conn       net.Conn
	conf       conf.ConfigYAML
	debug      bool
	SenderData *SenderPacket
	closeChan  chan struct{}
	sendChan   chan *Packet
	recvChan   chan *Packet
	active     bool
}

func NewClient(conf conf.ConfigYAML, debug bool) (*Client, error) {
	crt := conf.Security.GetCert()
	if crt == nil {
		return nil, errors.New("no certificate")
	}
	conn, err := tls.Dial("tcp", conf.Net.TargetAddr+":"+strconv.Itoa(int(conf.Net.TargetPort)), &tls.Config{
		Certificates: []tls.Certificate{*crt},
		RootCAs:      conf.Security.GetCertPool(),
	})
	if err != nil {
		return nil, err
	}
	return newClient(conf, conn, debug)
}

func newClient(cnf conf.ConfigYAML, conn net.Conn, debug bool) (*Client, error) {
	var sd *SenderPacket
	if cnf.GetMode() == conf.UnStore || cnf.GetMode() == conf.Backup {
		pk := &SenderPacket{
			Services:               &ServiceList{List: cnf.Services.List},
			RequestReboot:          cnf.TriggerReboot,
			RequestServiceStop:     cnf.Services.Stop,
			RequestServiceRestart:  cnf.Services.Restore,
			RequestServiceStartNew: cnf.Services.StartNew,
		}
		_, err := pk.WriteTo(conn)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		rv := &IngesterPacket{}
		_, err = rv.ReadFrom(conn)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
	} else {
		rv := &SenderPacket{}
		_, err := rv.ReadFrom(conn)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		pk := &IngesterPacket{}
		_, err = pk.WriteTo(conn)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		sd = rv
	}
	cl := &Client{
		conn:       conn,
		conf:       cnf,
		debug:      debug,
		SenderData: sd,
		closeChan:  make(chan struct{}),
		sendChan:   make(chan *Packet),
		recvChan:   make(chan *Packet),
		active:     true,
	}
	go func() {
		defer func() {
			cl.active = false
		}()
		for cl.active {
			select {
			case <-cl.closeChan:
				return
			case p := <-cl.sendChan:
				_, err := p.WriteTo(cl.conn)
				if err != nil {
					if cl.debug {
						log.Error(err)
					}
					cl.close(false)
					return
				}
			}
		}
	}()
	go func() {
		defer func() {
			cl.active = false
			close(cl.closeChan)
		}()
		for cl.active {
			select {
			case <-cl.closeChan:
				return
			default:
				p := &Packet{}
				_, err := p.ReadFrom(cl.conn)
				if err != nil {
					if cl.debug {
						log.Error(err)
					}
					cl.close(false)
					return
				}
				if p.Type == Finish {
					cl.close(false)
					return
				}
				select {
				case <-cl.closeChan:
					return
				case cl.recvChan <- p:
				}
			}
		}
	}()
	return cl, nil
}

func (c *Client) SendPacket(p *Packet) {
	if c.active {
		if p == nil {
			return
		}
		select {
		case c.sendChan <- p:
		case <-c.closeChan:
		}
	}
}

func (c *Client) ReceivePacket() *Packet {
	select {
	case p := <-c.recvChan:
		return p
	case <-c.closeChan:
	}
	return nil
}

func (c *Client) close(sendEnd bool) {
	defer func() {
		_ = c.conn.Close()
	}()
	if sendEnd {
		c.active = false
		select {
		case <-c.closeChan:
		}
		fp := &FinishPacket{}
		_, err := fp.WriteTo(c.conn)
		if err != nil && c.debug {
			log.Error(err)
		}
	} else {
		c.active = false
	}
}

func (c *Client) Close() {
	c.close(true)
}
