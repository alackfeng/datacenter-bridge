package datacenterbridge

import (
	"errors"
	"os"

	"github.com/alackfeng/datacenter-bridge/channel/quic"
	"github.com/alackfeng/datacenter-bridge/channel/websocket"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
	"gopkg.in/yaml.v3"
)

var ErrNoDiscovery = errors.New("no discovery")
var ErrNoServer = errors.New("no server")

// Configure -
type Configure struct {
	AppConfigure `yaml:"app" json:"app" comment:"应用配置"`
}

// NewConfigure -
func NewConfigure() *Configure {
	return &Configure{}
}

// LoadConfigure - load configure from file.
func LoadConfigure(configFile string) (*Configure, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config := &Configure{}
	if err := yaml.Unmarshal(file, config); err != nil {
		return nil, err
	}
	return config, nil
}

// Check - check configure.
func (c Configure) Check(server bool) error {
	if err := c.Discovery.check(); err != nil {
		return err
	}
	if !server {
		return nil
	}
	if err := c.Servers.check(); err != nil {
		return err
	}
	return nil
}

func (c Configure) Self() *discovery.Service {
	return &discovery.Service{
		Zone:    c.AppInfo.Zone,
		Service: c.AppInfo.Service,
		Id:      c.AppInfo.Id,
	}
}

type AppConfigure struct {
	AppInfo `yaml:",inline"`
	Mode    string `yaml:"mode" json:"mode" comment:"运行模式"`

	Log       logger.LogConfigure `yaml:"log" json:"log" comment:"日志配置"`
	Servers   ServerConfigure     `yaml:"servers" json:"servers" comment:"服务列表"`
	Discovery DiscoveryConfigure  `yaml:"discovery" json:"discovery" comment:"服务发现"`
}

// AppInfo - local app info.
type AppInfo struct {
	Zone    string `yaml:"zone" json:"zone" comment:"区域:cn-001"`
	Service string `yaml:"service" json:"service" comment:"服务类别:gw-dcb-service"`
	Id      string `yaml:"id" json:"id" comment:"服务Id:gw-node1"`
}

// NewSelf -
func NewSelf(zone, service, id string) *discovery.Service {
	return &discovery.Service{
		Zone:    zone,
		Service: service,
		Id:      id,
	}
}

// Register - etcd use.
func (c Configure) Register() *discovery.Service {
	host := c.Servers.Ws.Host
	if c.Servers.Quic.Up {
		host = c.Servers.Quic.Host
	} else if c.Servers.Wss.Up {
		host = c.Servers.Wss.Host
	}
	return &discovery.Service{
		Zone:    c.Zone,
		Service: c.Service,
		Id:      c.Id,
		Host:    host,
		Tag:     "primary",
	}
}

func (c Configure) String() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

// ServerConfigure -
type ServerConfigure struct {
	Ws   WebsocketConfigure  `yaml:"ws" json:"ws" comment:"Ws服务配置"`
	Wss  WebsocketsConfigure `yaml:"wss" json:"wss" comment:"Wss服务配置"`
	Quic QuicConfigure       `yaml:"quic" json:"quic" comment:"Quic服务配置"`
}

func (c ServerConfigure) check() error {
	if !c.Ws.Up && !c.Wss.Up && !c.Quic.Up {
		return ErrNoServer
	}
	return nil
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
	Up       bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host     string `yaml:"host" json:"host" comment:"ws://Ip:Port/bridge"`
	CertFile string `yaml:"certfile" json:"certfile" comment:"certfile"`
	KeyFile  string `yaml:"keyfile" json:"keyfile" comment:"keyfile"`
}

// To -
func (s WebsocketsConfigure) To() *websocket.WebsocketConfig {
	return websocket.NewWebsocketTlsConfig(s.Host, s.CertFile, s.KeyFile)
}

type QuicConfigure struct {
	Up       bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host     string `yaml:"host" json:"host" comment:"quic://Ip:Port"`
	CertFile string `yaml:"certfile" json:"certfile" comment:"certfile"`
	KeyFile  string `yaml:"keyfile" json:"keyfile" comment:"keyfile"`
}

// To -
func (s QuicConfigure) To() *quic.QuicConfig {
	return quic.NewQuicTlsConfig(s.Host, s.CertFile, s.KeyFile)
}

// DiscoveryConfigure -
type DiscoveryConfigure struct {
	Consul ConsulConfigure `yaml:"consul" json:"consul" comment:"Consul服务发现"`
	Etcd   EtcdConfigure   `yaml:"etcd" json:"etcd" comment:"Etcd服务发现"`
}

func (c *DiscoveryConfigure) check() error {
	if !c.Consul.Up && !c.Etcd.Up {
		return ErrNoDiscovery
	}
	return nil
}

// ConsulConfigure-
type ConsulConfigure struct {
	Up    bool   `yaml:"up" json:"up" comment:"是否启用"`
	Host  string `yaml:"host" json:"host" comment:"http://Ip:Port"`
	Token string `yaml:"token" json:"token" comment:"acl认证"`
}

type EtcdConfigure struct {
	Up         bool     `yaml:"up" json:"up" comment:"是否启用"`
	Endpoints  []string `yaml:"endpoints" json:"endpoints" comment:"[]Ip:Port"`
	Prefix     string   `yaml:"prefix" json:"prefix" comment:"service prefix /dcbridge."`
	GrantedTTL int64    `yaml:"ttl" json:"ttl" comment:"service granted ttl in seconds"`
}
