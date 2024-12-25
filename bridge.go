package datacenterbridge

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/channel/quic"
	"github.com/alackfeng/datacenter-bridge/channel/websocket"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
)

const channelChanCount = 10

// SWG -
var SWG = sync.WaitGroup{}

// DCenterBridge -
type DCenterBridge struct {
	ctx         context.Context
	done        chan bool
	config      Configure
	wsServer    *websocket.WebsocketServer
	quicServer  *quic.QuicServer
	discovery   discovery.Discovery // 服务发现.
	channels    sync.Map            // zone_service -> map[service_id]*channel.Channel.
	channelChan chan channel.Channel
}

var _ Datacenter = (*DCenterBridge)(nil)

// NewDCenterBridge -
func NewDCenterBridgeWithConfig(ctx context.Context, done chan bool, config *Configure) Datacenter {
	return &DCenterBridge{
		ctx:         ctx,
		done:        done,
		config:      *config,
		wsServer:    nil,
		quicServer:  nil,
		discovery:   nil,
		channels:    sync.Map{},
		channelChan: make(chan channel.Channel, channelChanCount),
	}
}

func NewDCenterBridge(ctx context.Context, done chan bool, config *Configure) Datacenter {
	return &DCenterBridge{
		ctx:         ctx,
		done:        done,
		config:      *config,
		wsServer:    nil,
		discovery:   discovery.NewConsulDiscovery(config.Discovery.Consul.Host),
		channelChan: make(chan channel.Channel, channelChanCount),
	}
}

// channelRead - 读取消息转发.
func (dc *DCenterBridge) channelRead(ch channel.Channel) {
	chInChan := ch.InChan()
	chDoneChan := ch.DoneChan()
	for {
		select {
		case <-dc.ctx.Done():
			logger.Warn("channelRead done.")
			return
		case <-chDoneChan:
			logger.Warn("channelRead closed.")
			dc.channels.Delete(ch.Key())
			return
		case data, ok := <-chInChan:
			if !ok {
				logger.Warn("channelRead closed.")
				return
			}
			logger.Debugf("channelRead channel: %+v, data: %s.", ch, string(data))
		}
	}
}

// ChannelsLoop -
func (dc *DCenterBridge) ChannelsLoop() error {
	logger.Warn("channelsLoop begin.")
	for {
		select {
		case <-dc.ctx.Done():
			logger.Warn("channelsLoop done.")
			return nil
		case ch, ok := <-dc.channelChan:
			if !ok {
				logger.Warn("channelsLoop closed.")
				return nil
			}
			go ch.ReadLoop()
			go ch.WriteLoop()
			go dc.channelRead(ch)

			logger.Debugf("channelsLoop channel: %+v", ch)
			if v, ok := dc.channels.Load(ch.Key()); ok {
				chs := v.([]channel.Channel)
				chs = append(chs, ch)
				dc.channels.Store(ch.Key(), chs)
			} else {
				dc.channels.Store(ch.Key(), []channel.Channel{ch})
			}
		}
	}
}

// ListenAndServe -
func (dc *DCenterBridge) ListenAndServe() error {
	d := dc.config.Discovery
	if d.Consul.Up {
		dc.discovery = discovery.NewConsulDiscovery(d.Consul.Host)
		logger.Infof("use discovery consul up<%v>, host<%v>", d.Consul.Up, d.Consul.Host)
	} else {
		logger.Error("no use discovery")
		return fmt.Errorf("discovery config error")
	}

	s := dc.config.Servers
	if s.Ws.Up {
		wsConfig := s.Ws.To()
		if wsConfig == nil {
			logger.Error("websocket server config error.")
			return fmt.Errorf("websocket server config error")
		}
		dc.wsServer = websocket.NewWebsocketServer(dc.config.Self(), wsConfig)
		SWG.Add(1)
		go func() {
			dc.wsServer.ListenAndServe(dc.ctx, dc.channelChan)
			SWG.Done()
		}()
		logger.Infof("use server websocket up<%v>, host<%v>", s.Ws.Up, wsConfig.Url())
	}
	if s.Quic.Up {
		dc.quicServer = quic.NewQuicServer(dc.config.Self(), s.Quic.To())
		SWG.Add(1)
		go func() {
			dc.quicServer.ListenAndServe(dc.ctx, dc.channelChan)
			SWG.Done()
		}()
		logger.Infof("use server quic up<%v>, host<%v>", s.Quic.Up, s.Quic.Host)
	}

	go dc.ChannelsLoop() // chan监听.
	return nil
}

// WaitQuit -
func (dc *DCenterBridge) WaitQuit() {
	go func() {
		logger.Warn("waiting quit...")
		SWG.Wait()
		dc.done <- true
		logger.Warn("waiting quit...done")
	}()
}

// selectService -
func (dc *DCenterBridge) selectService(zone, serviceName string) (*discovery.Service, error) {
	// 通过discovery获取service列表.
	services, err := dc.discovery.GetServices(dc.ctx, zone, serviceName)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("service not found")
	}
	// 随机选取一个Service.
	i := 1 //rand.Intn(len(services))
	service := services[i]
	logger.Debugf("select service: %d - %+v", i, service)
	return &service, nil
}

// CreateChannel - 创建桥通道.
func (dc *DCenterBridge) CreateChannel(zone, serviceName string) (channel.Channel, error) {
	if services, ok := dc.channels.Load(fmt.Sprintf("%s_%s", zone, serviceName)); ok {
		if chs, ok := services.([]channel.Channel); ok { // 已经存在连接, 直接选择.
			i := rand.Intn(len(chs))
			fmt.Println(">>> channel: ", i, chs[i].ID())
			return chs[i], nil
		}
	}
	peer, err := dc.selectService(zone, serviceName)
	if err != nil {
		return nil, err
	}
	switch scheme := peer.Scheme(); scheme {
	case "quic":
		quicChannel := quic.NewQuicClient(dc.config.Self(), peer)
		if err := quicChannel.Connect(dc.ctx); err != nil {
			return nil, err
		}
		dc.channelChan <- quicChannel // send to chan.
		logger.Debugf("connect to quic service: %+v, ok.", peer)
		return quicChannel, nil
	case "ws", "wss":
		wsChannel := websocket.NewWebsocketClient(dc.config.Self(), peer)
		if err := wsChannel.Connect(dc.ctx); err != nil {
			return nil, err
		}
		logger.Debugf("connect to service")
		dc.channelChan <- wsChannel // send to chan.
		logger.Debugf("connect to ws service: %+v, ok.", peer)
		return wsChannel, nil
	default:
		logger.Errorf("not support scheme: %s", scheme)
		return nil, fmt.Errorf("not support scheme: %s", scheme)
	}
}

// SendData -
func (dc *DCenterBridge) SendData(zone, serviceName string, data []byte) error {
	ch, err := dc.CreateChannel(zone, serviceName)
	if err != nil {
		return err
	}
	if err := ch.SendSafe(data); err != nil {
		return err
	}
	return nil
}
