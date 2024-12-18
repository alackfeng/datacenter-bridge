package channel

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
