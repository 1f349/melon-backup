package proxy

import (
	"github.com/1f349/melon-backup/comm"
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"net"
	"strconv"
	"sync"
)

type Listener struct {
	conf               conf.ConfigYAML
	active             bool
	conn               *comm.Client
	lstn               net.Listener
	connectionIDChan   chan int
	closeChan          chan struct{}
	connections        map[int]*Client
	connectionsLocker  *sync.RWMutex
	closeLocker        *sync.Mutex
	acceptCompleteChan chan bool
	acceptConnChan     chan net.Conn
}

func NewListener(conn *comm.Client, cnf conf.ConfigYAML) (*Listener, error) {
	tl, err := net.Listen("tcp", cnf.Net.GetProxyLocalAddr()+":"+strconv.Itoa(int(cnf.Net.GetProxyLocalPort())))
	if err != nil {
		return nil, err
	}
	ls := &Listener{
		conf:               cnf,
		active:             true,
		conn:               conn,
		lstn:               tl,
		connectionIDChan:   make(chan int),
		closeChan:          make(chan struct{}),
		connections:        make(map[int]*Client),
		connectionsLocker:  &sync.RWMutex{},
		closeLocker:        &sync.Mutex{},
		acceptCompleteChan: make(chan bool),
		acceptConnChan:     make(chan net.Conn),
	}
	go func() {
		defer ls.Close()
		for ls.active {
			cc, err := ls.lstn.Accept()
			if err != nil {
				if conf.Debug {
					log.Error(err)
				}
				ls.Close()
				return
			}
			if conf.Debug {
				log.Error("Accepted Client, awaiting connection: " + cc.RemoteAddr().String())
			}
			select {
			case <-ls.closeChan:
				_ = cc.Close()
				return
			case ls.acceptConnChan <- cc:
			}
			ls.conn.SendPacket(&comm.Packet{Type: comm.ConnectionStartRequest})
			select {
			case <-ls.closeChan:
				_ = cc.Close()
				return
			case <-ls.acceptCompleteChan:
			}
		}
	}()
	go func() {
		for ls.active {
			select {
			case <-ls.closeChan:
				return
			case cc := <-ls.acceptConnChan:
				select {
				case <-ls.closeChan:
					_ = cc.Close()
					return
				case cID := <-ls.connectionIDChan:
					if cID < 1 || ls.addClient(cc, cID) {
						_ = cc.Close()
						if conf.Debug {
							log.Error("Client not added!")
						}
					} else if conf.Debug {
						log.Error("Added Client: " + strconv.Itoa(cID))
						ls.conn.SendPacket(&comm.Packet{Type: comm.ConnectionSendStartRequest, ConnectionID: cID})
					}
					select {
					case <-ls.closeChan:
						return
					case ls.acceptCompleteChan <- true:
					}
				}
			}
		}
	}()
	go func() {
		defer ls.Close()
		for ls.active {
			p := conn.ReceivePacket()
			if p == nil {
				return
			}
			switch p.Type {
			case comm.ConnectionReset:
				select {
				case <-ls.closeChan:
					return
				case ls.connectionIDChan <- 0:
				default:
					if conf.Debug {
						log.Error("unexpected packet : ConnectionReset")
					}
				}
			case comm.ConnectionStarted:
				select {
				case <-ls.closeChan:
					return
				case ls.connectionIDChan <- p.ConnectionID:
				default:
					if conf.Debug {
						log.Error("unexpected packet : ConnectionStarted")
					}
					ls.conn.SendPacket(&comm.Packet{Type: comm.ConnectionClosed, ConnectionID: p.ConnectionID})
				}
			case comm.ConnectionData, comm.ConnectionClosed:
				if cc := ls.getClient(p.ConnectionID); cc != nil {
					cc.ReceivePacket(p)
				}
			}
		}
	}()
	return ls, nil
}

func (l *Listener) getClient(id int) *Client {
	l.connectionsLocker.RLock()
	defer l.connectionsLocker.RUnlock()
	if cc, ok := l.connections[id]; ok {
		return cc
	} else {
		return nil
	}
}

func (l *Listener) addClient(c net.Conn, id int) bool {
	if l.active {
		l.connectionsLocker.Lock()
		defer l.connectionsLocker.Unlock()
		if _, exs := l.connections[id]; exs {
			return true
		}
		nCl := newClient(l.conn, id, c, l.conf.Net.GetProxyBufferSize())
		nCl.StartSend()
		l.connections[id] = nCl
		go func() {
			select {
			case <-l.closeChan:
			case <-nCl.GetCloseChan():
			}
			l.connectionsLocker.Lock()
			defer l.connectionsLocker.Unlock()
			delete(l.connections, id)
		}()
		return false
	}
	return true
}

func (l *Listener) Close() {
	l.closeLocker.Lock()
	defer l.closeLocker.Unlock()
	if l.active {
		l.active = false
		_ = l.lstn.Close()
		l.connectionsLocker.RLock()
		defer l.connectionsLocker.RUnlock()
		for _, conn := range l.connections {
			conn.Close()
		}
	}
}
