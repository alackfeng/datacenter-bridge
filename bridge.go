package datacenterbridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/alackfeng/datacenter-bridge/channel/websocket"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
)

// SWG -
var SWG = sync.WaitGroup{}

// DCenterBridge -
type DCenterBridge struct {
	ctx       context.Context
	done      chan bool
	config    Configure
	wsServer  *websocket.WebsocketServer
	disConsul *discovery.ConsulDiscovery
	// channels sync.Map // websocket channels.
}

var _ Datacenter = (*DCenterBridge)(nil)

// NewDCenterBridge -
func NewDCenterBridgeWithConfig(ctx context.Context, done chan bool, config *Configure) Datacenter {
	return &DCenterBridge{
		ctx:       ctx,
		done:      done,
		config:    *config,
		wsServer:  nil,
		disConsul: nil,
	}
}

func NewDCenterBridge(ctx context.Context, done chan bool, discoveryUrl string) Datacenter {
	return &DCenterBridge{
		ctx:       ctx,
		done:      done,
		config:    Configure{},
		wsServer:  nil,
		disConsul: discovery.NewConsulDiscovery(discoveryUrl),
	}
}

// ListenAndServe -
func (dc *DCenterBridge) ListenAndServe() error {
	d := dc.config.Discovery
	if d.Consul.Up {
		dc.disConsul = discovery.NewConsulDiscovery(d.Consul.Host)
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
		dc.wsServer = websocket.NewWebsocketServer(wsConfig)
		SWG.Add(1)
		go func() {
			dc.wsServer.ListenAndServe(dc.ctx)
			SWG.Done()
		}()
		logger.Infof("use server websocket up<%v>, host<%v>", s.Ws.Up, wsConfig.Url())
	}
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

// CreateChannel -
func (dc *DCenterBridge) CreateChannel(zone, serviceName string) error {
	dc.disConsul.GetService(dc.ctx, zone, serviceName)

	wsChannel := websocket.NewWebsocketClient(websocket.NewWebsocketConfig("ws://127.0.0.1:8080/ws", "nil"))
	if err := wsChannel.Connect(); err != nil {
		fmt.Println("xxxx")
		return err
	}

	return nil
}
