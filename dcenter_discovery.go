package datacenterbridge

// DCenterDiscovery - 服务发现接口.
type DCenterDiscovery interface {
}

// dcenterDiscovery -
type dcenterDiscovery struct {
}

var _ DCenterDiscovery = (*dcenterDiscovery)(nil)

func NewDCenterDiscovery() DCenterDiscovery {
	return &dcenterDiscovery{}
}
