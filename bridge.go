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

// DCenterBridge - 区域桥通道.
type DCenterBridge struct {
	ctx         context.Context
	done        chan bool
	config      Configure
	wsServer    *websocket.WebsocketServer
	quicServer  *quic.QuicServer
	discovery   discovery.Discovery  // 服务发现.
	channels    sync.Map             // 区域通道列表: zone_service -> []channel.Channel.
	channelChan chan channel.Channel // 接受通道建立.
}

var _ Datacenter = (*DCenterBridge)(nil)

// NewDCenterBridge - server use.
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

// NewDCenterBridgeWithClient - client use.
func NewDCenterBridgeWithClient(ctx context.Context, done chan bool, self *discovery.Service, options ...interface{}) Datacenter {
	config := Configure{
		Zone:    self.Zone,
		Service: self.Service,
		Id:      self.Id,
	}
	for _, option := range options {
		switch v := option.(type) {
		case logger.LogConfigure:
			config.Log = v
		case *logger.LogConfigure:
			config.Log = *v
		case DiscoveryConfigure:
			config.Discovery = v
		case *DiscoveryConfigure:
			config.Discovery = *v
		case EtcdConfigure:
			config.Discovery.Etcd = v
		case *EtcdConfigure:
			config.Discovery.Etcd = *v
		case ConsulConfigure:
			config.Discovery.Consul = v
		case *ConsulConfigure:
			config.Discovery.Consul = *v
		default:
			logger.Warnf("NewDCenterBridgeWithClient unknown option: %+v.", v)
		}
	}
	return &DCenterBridge{
		ctx:         ctx,
		done:        done,
		config:      config,
		wsServer:    nil,
		quicServer:  nil,
		discovery:   discovery.NewConsulRegistry(config.Discovery.Consul.Host),
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
			dc.channels.Range(func(key, value interface{}) bool {
				chs := value.([]channel.Channel)
				for i, c := range chs {
					if c == ch {
						logger.Warn("channelRead closed, find it.")
						chs = append(chs[:i], chs[i+1:]...)
						dc.channels.Store(key, chs)
						return false
					}
				}
				return true
			})
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

func (dc *DCenterBridge) initDiscovery() error {
	d := dc.config.Discovery
	if d.Consul.Up {
		dc.discovery = discovery.NewConsulRegistry(d.Consul.Host)
		logger.Infof("use discovery consul up<%v>, host<%v>", d.Consul.Up, d.Consul.Host)
	} else if d.Etcd.Up {
		etcdDis := discovery.NewEtcdRegistry(d.Etcd.Endpoints, d.Etcd.Prefix, d.Etcd.GrantedTTL)
		if etcdDis == nil {
			logger.Error("etcd discovery err")
			return fmt.Errorf("discovery config error")
		}
		if err := etcdDis.Register(dc.ctx, *dc.config.Register()); err != nil {
			logger.Error("etcd register err")
			return fmt.Errorf("discovery config error")
		}
		etcdDis.Watch(dc.ctx)
		dc.discovery = etcdDis
		logger.Infof("use discovery etcd up<%v>, endpoints<%v>", d.Etcd.Up, d.Etcd.Endpoints)
	} else {
		logger.Error("no use discovery")
		return fmt.Errorf("discovery config error")
	}
	return nil
}

func (dc *DCenterBridge) initServers() error {
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
	if s.Wss.Up {
		wssConfig := s.Wss.To()
		if wssConfig == nil {
			logger.Error("websockets server config error.")
			return fmt.Errorf("websockets server config error")
		}
		dc.wsServer = websocket.NewWebsocketServer(dc.config.Self(), wssConfig)
		SWG.Add(1)
		go func() {
			dc.wsServer.ListenAndServe(dc.ctx, dc.channelChan)
			SWG.Done()
		}()
		logger.Infof("use server websockets up<%v>, host<%v>", s.Wss.Up, wssConfig.Url())
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
	return nil
}

// ListenAndServe -
func (dc *DCenterBridge) ListenAndServe() error {
	if err := dc.initDiscovery(); err != nil {
		return err
	}
	if err := dc.initServers(); err != nil {
		return err
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
