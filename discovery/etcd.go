package discovery

import "context"

// EtcdDiscovery -
type EtcdDiscovery struct {
}

// GetServices - implements Discovery.
func (e *EtcdDiscovery) GetServices(ctx context.Context, zone string, serviceName string) ([]Service, error) {
	panic("unimplemented")
}

var _ Discovery = (*EtcdDiscovery)(nil)

func NewEtcdDiscovery() Discovery {
	return &EtcdDiscovery{}
}
