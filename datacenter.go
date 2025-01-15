package datacenterbridge

import (
	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/discovery"
)

// GetChannelMsg - get bridge channel message.
type GetChannelMsg func(ch channel.Channel, data []byte)
type ClosedChannel func(ch channel.Channel)

// Datacenter - datacenter bridge interface.
type Datacenter interface {
	ListenAndServe() error // 启动服务监听.
	WaitQuit()             // 等待退出.

	ChannelsLoop(GetChannelMsg, ClosedChannel) error // client loop.

	DiscoveryServers(zone, service string) ([]discovery.Service, error) // 发现桥通道:区域|服务名称.

	CreateChannel(zone, service, id string) (channel.Channel, error) // 创建桥通道:区域|服务名称.
	DeleteChannel(zone, service, id string) error                    // 删除桥通道:区域|服务名称.
	SendData(zone, service string, data []byte) error                // 发送数据:区域|服务名称|数据.

	CreateChannelForTest(zone, service, id, host string) (channel.Channel, error) // for test, no use production.

	GetChannelList() ChannelList   // 获取桥通道列表.
	GetListenerList() ListenerList // 获取监听服务列表.

}
