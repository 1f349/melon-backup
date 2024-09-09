package conf

import (
	"net"
	"slices"
	"strings"
)

type NetYAML struct {
	TargetAddr         string              `yaml:"targetAddress"`
	TargetPort         uint16              `yaml:"targetPort"`
	TargetExpectedName string              `yaml:"targetExpectedName"`
	ListeningAddr      string              `yaml:"listeningAddress"`
	ListeningPort      uint16              `yaml:"listeningPort"`
	NameToIPsMapping   map[string][]string `yaml:"nameToIPsMapping"`
	ProxyLocalAddr     string              `yaml:"rsyncLocalAddr"`
	ProxyLocalPort     uint16              `yaml:"rsyncLocalPort"`
	ProxyBufferSize    uint32              `yaml:"proxyBufferSize"`
}

func (c NetYAML) GetTargetExpectedName() string {
	if c.TargetExpectedName == "" {
		return c.TargetAddr
	}
	return c.TargetExpectedName
}

func (c NetYAML) GetNameFromIP(ip string) string {
	for n, ips := range c.NameToIPsMapping {
		if slices.Contains(ips, ip) {
			return n
		}
	}
	addr, err := net.LookupAddr(ip)
	if err != nil && len(addr) > 0 {
		return strings.TrimRight(addr[0], ".")
	}
	return ip
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
