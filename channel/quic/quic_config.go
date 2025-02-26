package quic

import (
	"net/url"
	"time"
)

const defaultQueueSize = 100
const defaultBufferSize = 102400

// QuicConfig -
type QuicConfig struct {
	Host          string // localhost:4242.
	CertFile      string // certfile.
	KeyFile       string // keyfile.
	InChanCount   int
	OutChanCount  int
	HeartTimeoutS int
	HeartCount    int
}

// NewQuicConfig -
func NewQuicConfig(host string, queueSize int, bufferSize int) *QuicConfig {
	return &QuicConfig{
		Host:          host,
		InChanCount:   queueSize, // default 10.
		OutChanCount:  queueSize, // default 10.
		HeartTimeoutS: 30,
		HeartCount:    3,
	}
}

// NewQuicTlsConfig -
func NewQuicTlsConfig(host string, certFile, keyFile string, queueSize int, bufferSize int) *QuicConfig {
	c := NewQuicConfig(host, queueSize, bufferSize)
	c.CertFile = certFile
	c.KeyFile = keyFile
	return c
}

func (c QuicConfig) Addr() string {
	if u, err := url.Parse(c.Host); err != nil {
		return c.Host
	} else {
		return u.Host
	}
}

// MaxIdleTimeout -
func (c QuicConfig) MaxIdleTimeout() time.Duration {
	return time.Duration(c.HeartTimeoutS*c.HeartCount+10) * time.Second
}

// MaxIdleTimeout -
func (c QuicConfig) HandshakeIdleTimeout() time.Duration {
	return 10 * time.Second
}

// Keepalive -
func (c QuicConfig) Keepalive() time.Duration {
	return time.Duration(c.HeartTimeoutS) * time.Second
}
