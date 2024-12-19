package channel

import (
	"errors"
)

var (
	ErrDoneChannelClosed   = errors.New("done chan closed")
	ErrOutChannelFull      = errors.New("out chan full")
	ErrWriteNotFull        = errors.New("write not full")
	ErrMessageTypeNotMatch = errors.New("message type not match")
)

// Channel -
type Channel interface {
	ID() string
	Key() string
	DoneChan() chan struct{}
	InChan() chan []byte
	ReadLoop()
	WriteLoop()
	Close() error
	SendSafe(data []byte) error
}
