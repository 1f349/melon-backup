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
	debug             bool
	active            bool
	conn              *comm.Client
	newConnectionChan chan bool
	closeChan         chan struct{}
	connections       map[int]*Client
	connectionsLocker *sync.RWMutex
	closeLocker       *sync.Mutex
	cID               int
}

func NewMultiplexer(conn *comm.Client, cnf conf.ConfigYAML, debug bool) *Multiplexer {
	mx := &Multiplexer{
		conf:              cnf,
		debug:             debug,
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
					if debug {
						log.Error("Connection Requested...")
					}
					cc, err := net.Dial("tcp", cnf.Net.GetProxyLocalAddr()+":"+strconv.Itoa(int(cnf.Net.GetProxyLocalPort())))
					if err != nil {
						if debug {
							log.Error(err)
						}
						conn.SendPacket(&comm.Packet{Type: comm.ConnectionReset})
						break
					} else if debug {
						log.Error("Client connected!")
					}
					if mx.addClient(cc) {
						_ = cc.Close()
						if debug {
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
			if debug {
				log.Error("DBG : COMM_RCV : " + strconv.Itoa(int(p.Type)))
			}
			switch p.Type {
			case comm.ConnectionStartRequest:
				select {
				case <-mx.closeChan:
					return
				case mx.newConnectionChan <- true:
				default:
					if debug {
						log.Error("unexpected packet : ConnectionStartRequest")
					}
					mx.conn.SendPacket(&comm.Packet{Type: comm.ConnectionReset})
				}
			case comm.ConnectionData, comm.ConnectionClosed:
				go func() {
					mx.connectionsLocker.RLock()
					defer mx.connectionsLocker.RUnlock()
					select {
					case <-mx.closeChan:
					case <-mx.connections[p.ConnectionID].GetCloseChan():
					case mx.connections[p.ConnectionID].GetPacketIntake() <- p:
					}
				}()
			}
		}
	}()
	return mx
}

func (m *Multiplexer) addClient(c net.Conn) bool {
	if m.active {
		defer func() { m.cID++ }()
		m.connectionsLocker.Lock()
		defer m.connectionsLocker.Unlock()
		if _, exs := m.connections[m.cID]; exs {
			return true
		}
		nCl := newClient(m.conn, m.cID, c, m.conf.Net.GetProxyBufferSize(), m.debug)
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
