package proxy

import (
	"github.com/1f349/melon-backup/comm"
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"net"
	"strconv"
	"sync"
)

type Multiplexer struct {
	conf              conf.ConfigYAML
	active            bool
	conn              *comm.Client
	newConnectionChan chan bool
	closeChan         chan struct{}
	connections       map[int]*Client
	connectionsLocker *sync.RWMutex
	closeLocker       *sync.Mutex
	cID               int
}

func NewMultiplexer(conn *comm.Client, cnf conf.ConfigYAML) *Multiplexer {
	mx := &Multiplexer{
		conf:              cnf,
		active:            true,
		conn:              conn,
		newConnectionChan: make(chan bool),
		closeChan:         make(chan struct{}),
		connections:       make(map[int]*Client),
		connectionsLocker: &sync.RWMutex{},
		closeLocker:       &sync.Mutex{},
		cID:               1,
	}
	go func() {
		for mx.active {
			select {
			case <-mx.closeChan:
				return
			case x := <-mx.newConnectionChan:
				if x {
					if conf.Debug {
						log.Error("Connection Requested...")
					}
					cc, err := net.Dial("tcp", cnf.Net.GetProxyLocalAddr()+":"+strconv.Itoa(int(cnf.Net.GetProxyLocalPort())))
					if err != nil {
						if conf.Debug {
							log.Error(err)
						}
						conn.SendPacket(&comm.Packet{Type: comm.ConnectionReset})
						break
					} else if conf.Debug {
						log.Error("Client connected!")
					}
					if mx.addClient(cc) {
						_ = cc.Close()
						if conf.Debug {
							log.Error("Client not added!")
						}
					}
				}
			}
		}
	}()
	go func() {
		defer mx.Close()
		for mx.active {
			p := conn.ReceivePacket()
			if p == nil {
				return
			}
			switch p.Type {
			case comm.ConnectionStartRequest:
				select {
				case <-mx.closeChan:
					return
				case mx.newConnectionChan <- true:
				default:
					if conf.Debug {
						log.Error("unexpected packet : ConnectionStartRequest")
					}
					mx.conn.SendPacket(&comm.Packet{Type: comm.ConnectionReset})
				}
			case comm.ConnectionData, comm.ConnectionClosed, comm.ConnectionSendStartRequest:
				if cc := mx.getClient(p.ConnectionID); cc != nil {
					if p.Type == comm.ConnectionSendStartRequest {
						cc.StartSend()
					} else {
						cc.ReceivePacket(p)
					}
				}
			}
		}
	}()
	return mx
}

func (m *Multiplexer) getClient(id int) *Client {
	m.connectionsLocker.RLock()
	defer m.connectionsLocker.RUnlock()
	if cc, ok := m.connections[id]; ok {
		return cc
	} else {
		return nil
	}
}

func (m *Multiplexer) addClient(c net.Conn) bool {
	if m.active {
		defer func() { m.cID++ }()
		m.connectionsLocker.Lock()
		defer m.connectionsLocker.Unlock()
		if _, exs := m.connections[m.cID]; exs {
			return true
		}
		nCl := newClient(m.conn, m.cID, c, m.conf.Net.GetProxyBufferSize())
		m.connections[m.cID] = nCl
		go func() {
			select {
			case <-m.closeChan:
			case <-nCl.GetCloseChan():
			}
			m.connectionsLocker.Lock()
			defer m.connectionsLocker.Unlock()
			delete(m.connections, m.cID)
		}()
		m.conn.SendPacket(&comm.Packet{Type: comm.ConnectionStarted, ConnectionID: m.cID})
		return false
	}
	return true
}

func (m *Multiplexer) Close() {
	m.closeLocker.Lock()
	defer m.closeLocker.Unlock()
	if m.active {
		m.active = false
		close(m.closeChan)
		m.connectionsLocker.RLock()
		defer m.connectionsLocker.RUnlock()
		for _, conn := range m.connections {
			conn.Close()
		}
	}
}

func (m *Multiplexer) GetCloseWaiter() <-chan struct{} {
	return m.closeChan
}
