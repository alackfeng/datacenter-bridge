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
	InChanCount     int
	OutChanCount    int
	ReadBufferSize  int
	WriteBufferSize int
	ReadLimit       int64
	ReadDeadlineS   int
	HandshakeS      int // 握手时间s.
}

// NewWebsocketConfig -
func NewWebsocketConfig(host, prefix string) *WebsocketConfig {
	u, err := url.Parse(host)
	if err != nil {
		fmt.Println("url parse err: ", err.Error())
		return nil
	}
	u.Hostname()
	return &WebsocketConfig{
		Scheme:          u.Scheme,
		Ip:              u.Hostname(),
		Port:            u.Port(),
		Prefix:          prefix,
		InChanCount:     10,
		OutChanCount:    10,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		ReadLimit:       1024 * 1024,
		ReadDeadlineS:   60,
		HandshakeS:      5,
	}
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
	return time.Now().Add(time.Duration(c.ReadDeadlineS) * time.Second)
}
