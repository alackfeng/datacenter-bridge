package datacenterbridge

// DCenterBridge -
type DCenterBridge interface {
	Connect(string) error
}

// dcenterBridge -
type dcenterBridge struct {
}

var _ DCenterBridge = (*dcenterBridge)(nil)

// NewDCenterBridge -
func NewDCenterBridge() DCenterBridge {
	return &dcenterBridge{}
}

// Connect -
func (dc *dcenterBridge) Connect(string) error {
	return nil
}
