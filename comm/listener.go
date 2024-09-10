package comm

import (
	"crypto/tls"
	"errors"
	"github.com/1f349/melon-backup/conf"
	"net"
	"strconv"
	"strings"
)

type Listener struct {
	lstn   net.Listener
	conf   conf.ConfigYAML
	debug  bool
	active bool
}

func NewListener(conf conf.ConfigYAML, debug bool) (*Listener, error) {
	crt := conf.Security.GetCert()
	if crt == nil {
		return nil, errors.New("no certificate")
	}
	lstn, err := tls.Listen("tcp", conf.Net.ListeningAddr+":"+strconv.Itoa(int(conf.Net.ListeningPort)), &tls.Config{
		Certificates: []tls.Certificate{*crt},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    conf.Security.GetCertPool(),
		VerifyConnection: func(cs tls.ConnectionState) error {
			for _, nom := range conf.Net.RemoteAllowedNames {
				if err := cs.PeerCertificates[0].VerifyHostname(nom); err == nil {
					return nil
				}
			}
			return errors.New("Failed Remote Name Verification : " + strings.Join(cs.PeerCertificates[0].DNSNames, " "))
		},
	})
	if err != nil {
		return nil, err
	}
	return &Listener{
		lstn:   lstn,
		conf:   conf,
		debug:  debug,
		active: true,
	}, nil
}

func (l *Listener) Accept() (*Client, error) {
	if l.active {
		cc, err := l.lstn.Accept()
		if err != nil {
			return nil, err
		}
		return newClient(l.conf, cc, l.debug)
	}
	return nil, errors.New("listener closed")
}

func (l *Listener) Close() {
	l.active = false
	_ = l.lstn.Close()
}
