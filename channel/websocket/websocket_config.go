package websocket

import (
	"fmt"
	"net/url"
	"time"
)

// WebsocketConfig -
type WebsocketConfig struct {
	Scheme          string //
	Ip              string // Ip+Port.
	Port            string // Ip+Port.
	Prefix          string // bridge.
	CertFile        string // certfile.
	KeyFile         string // 	keyfile.
	InChanCount     int
	OutChanCount    int
	ReadBufferSize  int
	WriteBufferSize int
	ReadLimit       int64
	HeartTimeoutS   int // 心跳超时.
	HeartCount      int // 心跳次数.
	HandshakeS      int // 握手时间s.
}

// NewWebsocketConfig -
func NewWebsocketConfig(host string) *WebsocketConfig {
	u, err := url.Parse(host)
	if err != nil {
		fmt.Println("url parse err: ", err.Error())
		return nil
	}
	return &WebsocketConfig{
		Scheme:          u.Scheme,
		Ip:              u.Hostname(),
		Port:            u.Port(),
		Prefix:          u.Path,
		InChanCount:     10,
		OutChanCount:    10,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		ReadLimit:       1024 * 1024,
		HeartTimeoutS:   30,
		HeartCount:      3,
		HandshakeS:      5,
	}
}

// NewWebsocketTlsConfig -
func NewWebsocketTlsConfig(host string, certFile, keyFile string) *WebsocketConfig {
	c := NewWebsocketConfig(host)
	c.CertFile = certFile
	c.KeyFile = keyFile
	return c
}

// Host -
func (c WebsocketConfig) Host() string {
	return fmt.Sprintf("%s:%s", c.Ip, c.Port)
}

// Url - websocket request url.
func (c WebsocketConfig) Url() string {
	return fmt.Sprintf("%s://%s%s", c.Scheme, c.Host(), c.Prefix)
}

// HandshakeDeadline -
func (c WebsocketConfig) HandshakeDeadline() time.Duration {
	return time.Second * time.Duration(c.HandshakeS)
}

// ReadDeadline -
func (c WebsocketConfig) ReadDeadline() time.Time {
	return time.Now().Add(time.Duration(c.HeartTimeoutS*c.HeartCount+10) * time.Second)
}

// WriteDeadline -
func (c WebsocketConfig) WriteDeadline() time.Time {
	return time.Now().Add(time.Duration(c.HeartTimeoutS*c.HeartCount+10) * time.Second)
}

// Keepalive -
func (c WebsocketConfig) Keepalive() time.Duration {
	return time.Duration(c.HeartTimeoutS) * time.Second
}
