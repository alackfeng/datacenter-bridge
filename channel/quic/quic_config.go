package quic

import "time"

// QuicConfig -
type QuicConfig struct {
	Host          string // localhost:4242.
	InChanCount   int
	OutChanCount  int
	HeartTimeoutS int
	HeartCount    int
}

// NewQuicConfig -
func NewQuicConfig(host string) *QuicConfig {
	return &QuicConfig{
		Host:          host,
		InChanCount:   10,
		OutChanCount:  10,
		HeartTimeoutS: 30,
		HeartCount:    3,
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
