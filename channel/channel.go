package channel

import (
	"errors"

	"github.com/alackfeng/datacenter-bridge/discovery"
)

var (
	ErrDoneChannelClosed   = errors.New("done chan closed")
	ErrOutChannelFull      = errors.New("out chan full")
	ErrWriteNotFull        = errors.New("write not full")
	ErrMessageTypeNotMatch = errors.New("message type not match")
)

type ChannelInfo struct {
	Local discovery.Service `json:"local" comment:"本地信息"`
	Peer  discovery.Service `json:"peer" comment:"对端信息"`
}

// Channel -
type Channel interface {
	Self() discovery.Service
	Peer() discovery.Service
	String() string
	Info() ChannelInfo
	ID() string
	Key() string
	DoneChan() chan struct{}
	InChan() chan []byte
	ReadLoop()
	WriteLoop()
	Close() error
	SendSafe(data []byte) error
}
