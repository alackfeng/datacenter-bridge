package datacenterbridge

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/channel/quic"
	"github.com/alackfeng/datacenter-bridge/channel/websocket"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
	"github.com/alackfeng/datacenter-bridge/utils"
)

const channelChanCount = 10

var ErrChannelNotFound = errors.New("channel not found")

// SWG -
var SWG = sync.WaitGroup{}

// DCenterBridge - 区域桥通道.
type DCenterBridge struct {
	ctx         context.Context
	done        chan bool
	config      Configure
	wsServer    *websocket.WebsocketServer
	wssServer   *websocket.WebsocketServer
	quicServer  *quic.QuicServer
	discovery   discovery.Discovery  // 服务发现.
	channels    sync.Map             // 区域通道列表: zone_service -> []channel.Channel.
	channelChan chan channel.Channel // 接受通道建立.
}

var _ Datacenter = (*DCenterBridge)(nil)

// NewDCenterBridge -
func NewDCenterBridge(ctx context.Context, done chan bool, config *Configure) *DCenterBridge {
	dc := &DCenterBridge{
		ctx:         ctx,
		done:        done,
		config:      *config,
		wsServer:    nil,
		quicServer:  nil,
		discovery:   nil,
		channelChan: make(chan channel.Channel, channelChanCount),
	}
	return dc
}

// NewDCenterBridge - server use.
func NewDCenterBridgeWithServer(ctx context.Context, done chan bool, self AppInfo, options ...interface{}) Datacenter {
	config := &Configure{
		AppConfigure: AppConfigure{
			AppInfo: self,
			Mode:    "debug",
		},
	}
	for _, option := range options {
		switch v := option.(type) {
		case string:
			config.Mode = v
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
		case ServerConfigure:
			config.Servers = v
		case *ServerConfigure:
			config.Servers = *v
		case WebsocketConfigure:
			config.Servers.Ws = v
		case *WebsocketConfigure:
			config.Servers.Ws = *v
		case WebsocketsConfigure:
			config.Servers.Wss = v
		case *WebsocketsConfigure:
			config.Servers.Wss = *v
		default:
			logger.Warnf("NewDCenterBridgeWithServer unknown option: %+v.", v)
		}
	}
	if err := config.Check(true); err != nil {
		logger.Fatal(err.Error())
	}
	return NewDCenterBridge(ctx, done, config)
}

// NewDCenterBridgeWithClient - client use.
func NewDCenterBridgeWithClient(ctx context.Context, done chan bool, self AppInfo, options ...interface{}) Datacenter {
	config := &Configure{
		AppConfigure: AppConfigure{
			AppInfo: self,
			Mode:    "debug",
		},
	}
	for _, option := range options {
		switch v := option.(type) {
		case string:
			config.Mode = v
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
	if err := config.Check(false); err != nil {
		logger.Fatal(err.Error())
	}

	dc := NewDCenterBridge(ctx, done, config)
	if err := dc.initDiscovery(); err != nil {
		logger.Fatal(err.Error())
	}
	return dc
}

// channelRead - 读取消息转发.
func (dc *DCenterBridge) channelRead(ch channel.Channel, channelMsg GetChannelMsg, channelClosed ClosedChannel) {
	chInChan := ch.InChan()
	chDoneChan := ch.DoneChan()
	for {
		select {
		case <-dc.ctx.Done():
			logger.Warn("channelRead done.")
			return
		case <-chDoneChan:
			logger.Warn("channelRead closed.")
			key := ch.Key()
			v, ok := dc.channels.Load(key)
			if !ok {
				continue
			}
			chs := v.([]channel.Channel)
			for i, c := range chs {
				if c == ch {
					logger.Warn("channelRead closed, find it.")
					if channelClosed != nil {
						channelClosed(ch) // 通知用户关闭.
					}
					chs = slices.Delete(chs, i, i+1)
					if len(chs) == 0 {
						dc.channels.Delete(key) // 删除空列表.
					} else {
						dc.channels.Store(key, chs)
					}
					break
				}
			}
			return
		case data, ok := <-chInChan:
			if !ok {
				logger.Warn("channelRead closed.")
				return
			}
			channelMsg(ch, data)
			logger.Debugf("channelRead channel: %+v, data: %s.", ch, string(data))
		}
	}
}

// ChannelsLoop -
func (dc *DCenterBridge) ChannelsLoop(channelMsg GetChannelMsg, channelClosed ClosedChannel) error {
	logger.Info("channelsLoop begin.")
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
			go dc.channelRead(ch, channelMsg, channelClosed)

			logger.Debugf("channelsLoop channel: %+v", ch)
			if v, ok := dc.channels.Load(ch.Key()); ok {
				chs := v.([]channel.Channel)
				chs = append(chs, ch) // TODO: 是否需要判断存在否?, 支持重复连接.
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
		dc.wssServer = websocket.NewWebsocketServer(dc.config.Self(), wssConfig)
		SWG.Add(1)
		go func() {
			dc.wssServer.ListenAndServe(dc.ctx, dc.channelChan)
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
	// go dc.ChannelsLoop() // wait accept chan Channel监听.
	return nil
}

// WaitQuit -
func (dc *DCenterBridge) WaitQuit() {
	stop, stopCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer stopCancel()
	select {
	case <-stop.Done():
		logger.Warn("waiting quit...timeout 10s")
		stopCancel()
		return
	default:
		logger.Warn("waiting quit...")
		SWG.Wait()
		logger.Warn("waiting quit...done")
	}
	dc.done <- true
}

// ListenerList -
type ListenerList struct {
	Self    AppInfo  `json:"app" comment:"本地服务信息"`
	Listens []string `json:"listens" comment:"监听地址列表"`
}

func (l ListenerList) String() string {
	return utils.FormatJson(l)
}

func (dc *DCenterBridge) GetListenerList() ListenerList {
	l := ListenerList{
		Self:    dc.config.AppInfo,
		Listens: []string{},
	}
	if dc.wsServer != nil {
		l.Listens = append(l.Listens, dc.wsServer.ListenAddress())
	}
	if dc.wssServer != nil {
		l.Listens = append(l.Listens, dc.wssServer.ListenAddress())
	}
	if dc.quicServer != nil {
		l.Listens = append(l.Listens, dc.quicServer.ListenAddress())
	}
	return l
}

type ChannelList struct {
	Info map[string][]channel.ChannelInfo `json:"info" comment:"桥通道服务列表"`
}

func (l ChannelList) String() string {
	return utils.FormatJson(l)
}

// GetChannelList -
func (dc *DCenterBridge) GetChannelList() ChannelList {
	cl := ChannelList{
		Info: make(map[string][]channel.ChannelInfo),
	}
	keys := []string{}
	dc.channels.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		chs := value.([]channel.Channel)
		var info []channel.ChannelInfo
		for _, ch := range chs {
			info = append(info, ch.Info())
		}
		cl.Info[key.(string)] = info
		return true
	})
	return cl
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
	i := rand.Intn(len(services))
	service := services[i]
	logger.Debugf("select service: %d - %+v", i, service)
	return &service, nil
}

// connectTo - 连接到peer, 返回通道.
func (dc *DCenterBridge) connectTo(peer *discovery.Service) (channel.Channel, error) {
	switch scheme := peer.Scheme(); scheme {
	case "quic":
		quicChannel := quic.NewQuicClient(dc.config.Self(), peer)
		if err := quicChannel.Connect(dc.ctx); err != nil {
			return nil, err
		}
		logger.Debugf("connect to quic service: %+v, ok.", peer)
		return quicChannel, nil
	case "ws", "wss":
		wsChannel := websocket.NewWebsocketClient(dc.config.Self(), peer)
		if err := wsChannel.Connect(dc.ctx); err != nil {
			return nil, err
		}
		logger.Debugf("connect to service")
		logger.Debugf("connect to ws service: %+v, ok.", peer)
		return wsChannel, nil
	default:
		logger.Errorf("not support scheme: %s", scheme)
		return nil, fmt.Errorf("not support scheme: %s", scheme)
	}
}

func (dc *DCenterBridge) CreateChannelForTest(zone, serviceName, id, host string) (channel.Channel, error) {
	peer := &discovery.Service{
		Zone:    zone,
		Service: serviceName,
		Id:      id,
		Host:    host,
		Tag:     "primary",
	}
	if ch, err := dc.connectTo(peer); err != nil {
		return nil, err
	} else {
		dc.channelChan <- ch // send to chan.
		return ch, nil
	}
}

func (dc *DCenterBridge) DeleteChannel(zone, serviceName string) error {
	key := fmt.Sprintf("%s_%s", zone, serviceName)
	if v, ok := dc.channels.Load(key); ok {
		chs := v.([]channel.Channel)
		for _, ch := range chs {
			ch.Close()
			break
		}
	} else {
		return ErrChannelNotFound
	}
	return nil
}

// CreateChannel - 创建桥通道.
func (dc *DCenterBridge) CreateChannel(zone, serviceName string) (channel.Channel, error) {
	if services, ok := dc.channels.Load(fmt.Sprintf("%s_%s", zone, serviceName)); ok {
		if chs, ok := services.([]channel.Channel); ok { // 已经存在连接, 直接选择.
			i := rand.Intn(len(chs))
			logger.Debugf(">>> get local cached channel: %d, %s", i, chs[i].String())
			return chs[i], nil
		}
	}
	peer, err := dc.selectService(zone, serviceName)
	if err != nil {
		return nil, err
	}
	if ch, err := dc.connectTo(peer); err != nil {
		return nil, err
	} else {
		dc.channelChan <- ch // send to chan.
		return ch, nil
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
