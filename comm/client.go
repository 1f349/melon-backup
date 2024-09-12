package comm

import (
	"crypto/tls"
	"errors"
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"net"
	"strconv"
	"time"
)

type Client struct {
	conn           net.Conn
	conf           conf.ConfigYAML
	debug          bool
	SenderData     *SenderPacket
	closeChan      chan struct{}
	sendChan       chan *Packet
	recvChan       chan *Packet
	active         bool
	pureSocketMode bool
}

func NewClient(conf conf.ConfigYAML, debug bool) (*Client, error) {
	crt := conf.Security.GetCert()
	if crt == nil {
		return nil, errors.New("no certificate")
	}
	conn, err := tls.Dial("tcp", conf.Net.TargetAddr+":"+strconv.Itoa(int(conf.Net.TargetPort)), &tls.Config{
		Certificates: []tls.Certificate{*crt},
		RootCAs:      conf.Security.GetCertPool(),
		ServerName:   conf.Net.GetTargetExpectedName(),
	})
	if err != nil {
		return nil, err
	}
	return newClient(conf, conn, debug)
}

func newClient(cnf conf.ConfigYAML, conn net.Conn, debug bool) (*Client, error) {
	var sd *SenderPacket
	if cnf.GetMode() == conf.UnStore || cnf.GetMode() == conf.Backup {
		if debug {
			log.Error("Sending Sender Packet...")
		}
		pk := &SenderPacket{
			Mode:                   cnf.GetMode().ToInt(),
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
		if debug {
			log.Error("Waiting for Ingester Packet...")
		}
		rv := &IngesterPacket{}
		_, err = rv.ReadFrom(conn)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		sd = &SenderPacket{Mode: rv.Mode}
	} else {
		if debug {
			log.Error("Waiting for Sender Packet...")
		}
		rv := &SenderPacket{}
		_, err := rv.ReadFrom(conn)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		if debug {
			log.Error("Sending Ingester Packet...")
		}
		pk := &IngesterPacket{Mode: cnf.GetMode().ToInt()}
		_, err = pk.WriteTo(conn)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		sd = rv
	}
	return &Client{
		conn:           conn,
		conf:           cnf,
		debug:          debug,
		SenderData:     sd,
		closeChan:      make(chan struct{}),
		sendChan:       make(chan *Packet),
		recvChan:       make(chan *Packet),
		active:         false,
		pureSocketMode: false,
	}, nil
}

func (c *Client) ActivateWithPacketProcessing() {
	if c.active {
		return
	}
	c.active = true
	go func() {
		var kAlive *time.Ticker = nil
		if c.conf.Net.KeepAliveTime > time.Millisecond {
			kAlive = time.NewTicker(c.conf.Net.KeepAliveTime)
		}
		pk := &Packet{Type: ConnectionKeepAlive}
		defer func() {
			c.active = false
			if kAlive != nil {
				kAlive.Stop()
			}
		}()
		for c.active {
			select {
			case <-c.closeChan:
				return
			case p := <-c.sendChan:
				_, err := p.WriteTo(c.conn)
				if err != nil {
					if c.debug {
						log.Error(err)
					}
					c.close(false)
					return
				}
			case <-kAlive.C:
				_, err := pk.WriteTo(c.conn)
				if err != nil {
					if c.debug {
						log.Error(err)
					}
					c.close(false)
					return
				}
				kAlive.Reset(c.conf.Net.KeepAliveTime)
			}
		}
	}()
	go func() {
		defer func() {
			c.active = false
			close(c.closeChan)
		}()
		for c.active {
			select {
			case <-c.closeChan:
				return
			default:
				p := &Packet{}
				_, err := p.ReadFrom(c.conn)
				if err != nil {
					if c.debug {
						log.Error(err)
					}
					c.close(false)
					return
				}
				if p.Type == Finish {
					c.close(false)
					return
				}
				if p.Type != ConnectionKeepAlive {
					select {
					case <-c.closeChan:
						return
					case c.recvChan <- p:
					}
				}
			}
		}
	}()
}

func (c *Client) ActivateForPureConnection() net.Conn {
	if c.active {
		return nil
	}
	c.active = true
	c.pureSocketMode = true
	return c.conn
}

func (c *Client) SendPacket(p *Packet) {
	if c.pureSocketMode {
		return
	}
	if c.active {
		if p == nil {
			return
		}
		select {
		case c.sendChan <- p:
			if conf.Debug {
				log.Error("DBG : COMM_SND : " + strconv.Itoa(int(p.Type)) + " : [" + strconv.Itoa(p.ConnectionID) + "]")
			}
		case <-c.closeChan:
		}
	}
}

func (c *Client) ReceivePacket() *Packet {
	if c.pureSocketMode {
		return nil
	}
	select {
	case p := <-c.recvChan:
		if conf.Debug {
			log.Error("DBG : COMM_RCV : " + strconv.Itoa(int(p.Type)) + " : [" + strconv.Itoa(p.ConnectionID) + "]")
		}
		return p
	case <-c.closeChan:
	}
	return nil
}

func (c *Client) close(sendEnd bool) {
	defer func() {
		_ = c.conn.Close()
	}()
	if sendEnd && !c.pureSocketMode {
		c.active = false
		fp := &FinishPacket{}
		_, err := fp.WriteTo(c.conn)
		if err != nil && c.debug {
			log.Error(err)
		}
		select {
		case <-c.closeChan:
		}
	} else {
		c.active = false
	}
}

func (c *Client) Close() {
	c.close(true)
}
