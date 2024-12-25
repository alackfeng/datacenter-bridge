package datacenterbridge

import (
	"github.com/alackfeng/datacenter-bridge/channel/quic"
	"github.com/alackfeng/datacenter-bridge/channel/websocket"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
)

// Configure -
type Configure struct {
	Zone      string              `yaml:"zone" json:"zone" comment:"区域:us-001"`
	Service   string              `yaml:"service" json:"service" comment:"服务类别:dc-bridge"`
	Id        string              `yaml:"id" json:"id" comment:"服务Id:gw-dc-bridge-node1"`
	Log       logger.LogConfigure `yaml:"log" json:"log" comment:"日志配置"`
	Servers   ServerConfigure     `yaml:"servers" json:"servers" comment:"服务列表"`
	Discovery DiscoveryConfigure  `yaml:"discovery" json:"discovery" comment:"服务发现"`
}

// NewConfigure -
func NewConfigure() *Configure {
	return &Configure{}
}

func (c Configure) Self() *discovery.Service {
	return &discovery.Service{
		Zone:    c.Zone,
		Service: c.Service,
		Id:      c.Id,
	}
}

// Register - etcd use.
func (c Configure) Register() *discovery.Service {
	host := c.Servers.Ws.Host
	if c.Servers.Quic.Up {
		host = c.Servers.Quic.Host
	}
	return &discovery.Service{
		Zone:    c.Zone,
		Service: c.Service,
		Id:      c.Id,
		Host:    host,
		Tag:     "primary",
	}
}

// ServerConfigure -
type ServerConfigure struct {
	Ws   WebsocketConfigure  `yaml:"ws" json:"ws" comment:"Ws服务配置"`
	Wss  WebsocketsConfigure `yaml:"wss" json:"wss" comment:"Wss服务配置"`
	Quic QuicConfigure       `yaml:"quic" json:"quic" comment:"Quic服务配置"`
}

// WebsocketConfigure -
type WebsocketConfigure struct {
	Up   bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host string `yaml:"host" json:"host" comment:"ws://Ip:Port/bridge"`
	// Prefix string `yaml:"prefix" json:"prefix" comment:"uri prefix"`
}

// To -
func (s WebsocketConfigure) To() *websocket.WebsocketConfig {
	return websocket.NewWebsocketConfig(s.Host)
}

// WebsocketsConfigure -
type WebsocketsConfigure struct {
	Up     bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host   bool   `yaml:"host" json:"host" comment:"Ip:Port"`
	Prefix bool   `yaml:"prefix" json:"prefix" comment:"uri prefix"`
	CaCert string `yaml:"cacert" json:"cacert" comment:"cacert"`
	CaKey  string `yaml:"cakey" json:"cakey" comment:"cakey"`
}

type QuicConfigure struct {
	Up   bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host string `yaml:"host" json:"host" comment:"quic://Ip:Port"`
}

// To -
func (s QuicConfigure) To() *quic.QuicConfig {
	return quic.NewQuicConfig(s.Host)
}

// DiscoveryConfigure -
type DiscoveryConfigure struct {
	Consul ConsulConfigure `yaml:"consul" json:"consul" comment:"Consul服务发现"`
	Etcd   EtcdConfigure   `yaml:"etcd" json:"etcd" comment:"Etcd服务发现"`
}

// ConsulConfigure-
type ConsulConfigure struct {
	Up   bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host string `yaml:"host" json:"host" comment:"http://Ip:Port"`
}

type EtcdConfigure struct {
	Up         bool     `yaml:"up" json:"up" comment:"是否启用"`
	Endpoints  []string `yaml:"endpoints" json:"endpoints" comment:"[]Ip:Port"`
	Prefix     string   `yaml:"prefix" json:"prefix" comment:"service prefix"`
	GrantedTTL int64    `yaml:"ttl" json:"ttl" comment:"service granted ttl in seconds"`
}
