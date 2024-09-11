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
	conf              conf.ConfigYAML
	debug             bool
	active            bool
	conn              *comm.Client
	lstn              net.Listener
	connectionIDChan  chan int
	closeChan         chan struct{}
	connections       map[int]*Client
	connectionsLocker *sync.RWMutex
	closeLocker       *sync.Mutex
}

func NewListener(conn *comm.Client, cnf conf.ConfigYAML, debug bool) (*Listener, error) {
	tl, err := net.Listen("tcp", cnf.Net.GetProxyLocalAddr()+":"+strconv.Itoa(int(cnf.Net.GetProxyLocalPort())))
	if err != nil {
		return nil, err
	}
	ls := &Listener{
		conf:              cnf,
		debug:             debug,
		active:            true,
		conn:              conn,
		lstn:              tl,
		connectionIDChan:  make(chan int),
		closeChan:         make(chan struct{}),
		connections:       make(map[int]*Client),
		connectionsLocker: &sync.RWMutex{},
		closeLocker:       &sync.Mutex{},
	}
	go func() {
		defer ls.Close()
		for ls.active {
			cc, err := ls.lstn.Accept()
			if err != nil {
				if debug {
					log.Error(err)
				}
				ls.Close()
				return
			}
			if debug {
				log.Error("Accepted Client, awaiting connection: " + cc.RemoteAddr().String())
			}
			ls.conn.SendPacket(&comm.Packet{Type: comm.ConnectionStartRequest})
			select {
			case <-ls.closeChan:
				_ = cc.Close()
				return
			case cID := <-ls.connectionIDChan:
				if cID < 1 || ls.addClient(cc, cID) {
					_ = cc.Close()
					if debug {
						log.Error("Client not added!")
					}
				} else if debug {
					log.Error("Added Client: " + strconv.Itoa(cID))
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
					if debug {
						log.Error("unexpected packet : ConnectionReset")
					}
				}
			case comm.ConnectionStarted:
				select {
				case <-ls.closeChan:
					return
				case ls.connectionIDChan <- p.ConnectionID:
				default:
					if debug {
						log.Error("unexpected packet : ConnectionStarted")
					}
					ls.conn.SendPacket(&comm.Packet{Type: comm.ConnectionClosed, ConnectionID: p.ConnectionID})
				}
			case comm.ConnectionData, comm.ConnectionClosed:
				go func() {
					ls.connectionsLocker.RLock()
					defer ls.connectionsLocker.RUnlock()
					select {
					case <-ls.closeChan:
					case <-ls.connections[p.ConnectionID].GetCloseChan():
					case ls.connections[p.ConnectionID].GetPacketIntake() <- p:
					}
				}()
			}
		}
	}()
	return ls, nil
}

func (l *Listener) addClient(c net.Conn, id int) bool {
	if l.active {
		l.connectionsLocker.Lock()
		defer l.connectionsLocker.Unlock()
		if _, exs := l.connections[id]; exs {
			return true
		}
		nCl := newClient(l.conn, id, c, l.conf.Net.GetProxyBufferSize(), l.debug)
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
