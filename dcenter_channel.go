package datacenterbridge

type DCenterChannel interface {
}

// dcenterChannel -
type dcenterChannel struct {
}

var _ DCenterChannel = (*dcenterChannel)(nil)
