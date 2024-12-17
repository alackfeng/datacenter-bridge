package datacenterbridge

// Datacenter - datacenter bridge interface.
type Datacenter interface {
	ListenAndServe() error                    // 启动服务监听.
	WaitQuit()                                // 等待退出.
	CreateChannel(zone, service string) error // 创建桥通道:区域|服务名称.
}
