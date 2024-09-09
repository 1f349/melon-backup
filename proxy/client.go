package proxy

import (
	"github.com/1f349/melon-backup/comm"
	"github.com/charmbracelet/log"
	"net"
)

type Client struct {
	conn         *comm.Client
	conID        int
	connTCP      net.Conn
	packetIntake chan *comm.Packet
	closeChan    chan struct{}
	active       bool
}

func newClient(conn *comm.Client, conID int, connTCP net.Conn, buffSize uint32, debug bool) *Client {
	cl := &Client{
		conn:         conn,
		conID:        conID,
		connTCP:      connTCP,
		closeChan:    make(chan struct{}),
		packetIntake: make(chan *comm.Packet),
		active:       true,
	}
	go func() {
		defer func() {
			cl.active = false
			close(cl.closeChan)
		}()
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
			select {
			case <-cl.closeChan:
				return
			case p := <-cl.packetIntake:
				if p.Type != comm.ConnectionData {
					cl.close(false)
					return
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

func (c *Client) close(sendEnd bool) {
	defer func() {
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

func (c *Client) GetPacketIntake() chan<- *comm.Packet {
	return c.packetIntake
}
