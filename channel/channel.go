package channel

// Channel -
type Channel interface {
	Read()
	Write()
	Colse() error
}
