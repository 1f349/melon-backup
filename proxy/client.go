package proxy

import (
	"encoding/base64"
	"github.com/1f349/melon-backup/comm"
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
	"net"
	"strconv"
)

type Client struct {
	conn            *comm.Client
	conID           int
	connTCP         net.Conn
	packetQueue     *utils.Queue[*comm.Packet]
	closeChan       chan struct{}
	notifySendStart chan bool
	active          bool
}

func newClient(conn *comm.Client, conID int, connTCP net.Conn, buffSize uint32, debug bool) *Client {
	cl := &Client{
		conn:            conn,
		conID:           conID,
		connTCP:         connTCP,
		closeChan:       make(chan struct{}),
		active:          true,
		notifySendStart: make(chan bool),
		packetQueue:     utils.NewQueue[*comm.Packet](),
	}
	go func() {
		defer func() {
			cl.active = false
			close(cl.closeChan)
			cl.packetQueue.StartUnBlocking()
			cl.packetQueue.Clear()
		}()
		awaitSendStart := true
		for cl.active && awaitSendStart {
			select {
			case <-cl.closeChan:
				return
			case <-cl.notifySendStart:
				awaitSendStart = false
			}
		}
		for cl.active {
			buff := make([]byte, buffSize)
			select {
			case <-cl.closeChan:
				return
			default:
				br, err := cl.connTCP.Read(buff)
				if err != nil {
					if debug {
						log.Error(err)
					}
					cl.close(true)
					return
				}
				if debug {
					log.Error("DBG : TCP (" + strconv.Itoa(br) + ")-> COM : " + strconv.Itoa(cl.conID))
					log.Error("DBG : " + strconv.Itoa(cl.conID) + ": -> : " + base64.StdEncoding.EncodeToString(buff[:br]))
				}
				p := &comm.Packet{
					Type:         comm.ConnectionData,
					ConnectionID: cl.conID,
					Data:         buff[:br],
				}
				cl.conn.SendPacket(p)
			}
		}
	}()
	go func() {
		defer func() {
			cl.active = false
		}()
		for cl.active {
			p := cl.packetQueue.Dequeue()
			if p != nil {
				if p.Type != comm.ConnectionData {
					cl.close(false)
					return
				}
				if debug {
					log.Error("DBG : TCP <-(" + strconv.Itoa(len(p.Data)) + ") COM : " + strconv.Itoa(cl.conID))
					log.Error("DBG : " + strconv.Itoa(cl.conID) + ": <- : " + base64.StdEncoding.EncodeToString(p.Data))
				}
				_, err := cl.connTCP.Write(p.Data)
				if err != nil {
					if debug {
						log.Error(err)
					}
					cl.close(true)
					return
				}
			}
		}
	}()
	return cl
}

func (c *Client) StartSend() {
	if c.active {
		select {
		case <-c.closeChan:
		case c.notifySendStart <- true:
		}
	}
}

func (c *Client) close(sendEnd bool) {
	defer func() {
		if conf.Debug {
			log.Error("DBG : CLOSED : " + strconv.Itoa(c.conID))
		}
		_ = c.connTCP.Close()
		if sendEnd {
			p := &comm.Packet{
				Type:         comm.ConnectionClosed,
				ConnectionID: c.conID,
			}
			c.conn.SendPacket(p)
		}
	}()
	c.active = false
}

func (c *Client) Close() {
	c.close(true)
}

func (c *Client) GetCloseChan() <-chan struct{} {
	return c.closeChan
}

func (c *Client) ReceivePacket(p *comm.Packet) {
	if c.active {
		c.packetQueue.Enqueue(p)
	}
}
