package conf

import "time"

type NetYAML struct {
	TargetAddr         string        `yaml:"targetAddress"`
	TargetPort         uint16        `yaml:"targetPort"`
	TargetExpectedName string        `yaml:"targetExpectedName"`
	ListeningAddr      string        `yaml:"listeningAddress"`
	ListeningPort      uint16        `yaml:"listeningPort"`
	RemoteAllowedNames []string      `yaml:"remoteAllowedNames"`
	ProxyLocalAddr     string        `yaml:"rsyncLocalAddr"`
	ProxyLocalPort     uint16        `yaml:"rsyncLocalPort"`
	ProxyBufferSize    uint32        `yaml:"proxyBufferSize"`
	KeepAliveTime      time.Duration `yaml:"keepAliveTime"`
}

func (c NetYAML) GetTargetExpectedName() string {
	if c.TargetExpectedName == "" {
		return c.TargetAddr
	}
	return c.TargetExpectedName
}

func (c NetYAML) GetProxyLocalAddr() string {
	if c.ProxyLocalAddr == "" {
		return "127.0.0.1"
	}
	return c.ProxyLocalAddr
}

func (c NetYAML) GetProxyLocalPort() uint16 {
	if c.ProxyLocalPort == 0 {
		c.ProxyLocalPort = 873
	}
	return c.ProxyLocalPort
}

func (c NetYAML) GetProxyBufferSize() uint32 {
	if c.ProxyBufferSize < 256 {
		return 256
	}
	return c.ProxyBufferSize
}
