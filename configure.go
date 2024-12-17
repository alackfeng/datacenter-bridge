package datacenterbridge

import (
	"github.com/alackfeng/datacenter-bridge/channel/websocket"
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

// Node -
type Node struct {
	Zone    string `yaml:"zone" json:"zone" comment:"区域:us-001"`
	Service string `yaml:"service" json:"service" comment:"服务类别:dc-bridge"`
	Id      string `yaml:"id" json:"id" comment:"服务Id:gw-dc-bridge-node1"`
}

// ServerConfigure -
type ServerConfigure struct {
	Ws  WebsocketConfigure  `yaml:"ws" json:"ws" comment:"Ws服务配置"`
	Wss WebsocketsConfigure `yaml:"wss" json:"wss" comment:"Wss服务配置"`
}

// WebsocketConfigure -
type WebsocketConfigure struct {
	Up     bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host   string `yaml:"host" json:"host" comment:"ws://Ip:Port"`
	Prefix string `yaml:"prefix" json:"prefix" comment:"uri prefix"`
}

// To -
func (s WebsocketConfigure) To() *websocket.WebsocketConfig {
	return websocket.NewWebsocketConfig(s.Host, s.Prefix)
}

// WebsocketsConfigure -
type WebsocketsConfigure struct {
	Up     bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host   bool   `yaml:"host" json:"host" comment:"Ip:Port"`
	Prefix bool   `yaml:"prefix" json:"prefix" comment:"uri prefix"`
	CaCert string `yaml:"cacert" json:"cacert" comment:"cacert"`
	CaKey  string `yaml:"cakey" json:"cakey" comment:"cakey"`
}

// DiscoveryConfigure -
type DiscoveryConfigure struct {
	Consul ConsulConfigure `yaml:"consul" json:"consul" comment:"Consul服务发现"`
}

// ConsulConfigure-
type ConsulConfigure struct {
	Up   bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host string `yaml:"host" json:"host" comment:"http://Ip:Port"`
}
