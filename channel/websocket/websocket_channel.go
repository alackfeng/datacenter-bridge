package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
	"github.com/gorilla/websocket"
)

// WebsocketChannel - websocket通道对象, 进行收发操作.
type WebsocketChannel struct {
	self        discovery.Service // 身份标识.
	peer        discovery.Service // 对端身份.
	config      WebsocketConfig
	conn        *websocket.Conn
	closeOnce   sync.Once
	doneChan    chan struct{}
	inChan      chan []byte
	outChan     chan []byte
	isConnected bool
	isClient    bool
}

var _ channel.Channel = (*WebsocketChannel)(nil)

// newWebsocketServerChannel -
func newWebsocketServerChannel(self *discovery.Service, peer *discovery.Service, config *WebsocketConfig) *WebsocketChannel {
	w := newWebsocketChannel(self, peer, config, false)
	return w
}

// newWebsocketClientChannel -
func newWebsocketClientChannel(self *discovery.Service, peer *discovery.Service) *WebsocketChannel {
	return newWebsocketChannel(self, peer, NewWebsocketConfig(peer.Host), true)
}

// newWebsocketChannel -
func newWebsocketChannel(self *discovery.Service, peer *discovery.Service, config *WebsocketConfig, isClient bool) *WebsocketChannel {
	return &WebsocketChannel{
		self:        *self,
		peer:        *peer,
		config:      *config,
		isClient:    isClient,
		isConnected: false,
		doneChan:    make(chan struct{}),
		inChan:      make(chan []byte, config.InChanCount),
		outChan:     make(chan []byte, config.OutChanCount),
	}
}

func (c *WebsocketChannel) init(conn *websocket.Conn) *WebsocketChannel {
	c.conn = conn
	c.isConnected = true
	c.self.Host = c.conn.LocalAddr().String()
	return c
}

// ID - service.Id.
func (c *WebsocketChannel) ID() string {
	return c.peer.Id
}

// Key - zone_service.
func (c *WebsocketChannel) Key() string {
	return c.peer.Key()
}

func (c *WebsocketChannel) Self() discovery.Service {
	return c.self
}
func (c *WebsocketChannel) Peer() discovery.Service {
	return c.peer
}
func (c *WebsocketChannel) Info() channel.ChannelInfo {
	return channel.ChannelInfo{
		Local: c.self,
		Peer:  c.peer,
	}
}

// String -
func (c *WebsocketChannel) String() string {
	return fmt.Sprintf("[%s:%s:%s]%s=>[%s:%s:%s]%s",
		c.self.Service, c.self.Zone, c.self.Id, c.self.Host,
		c.peer.Service, c.peer.Zone, c.peer.Id, c.peer.Host)
}

// DoneChan -
func (c *WebsocketChannel) DoneChan() chan struct{} {
	return c.doneChan
}

// InChan -
func (c *WebsocketChannel) InChan() chan []byte {
	return c.inChan
}

// ReadLoop -
func (c *WebsocketChannel) ReadLoop() {
	logger.Debugf("websocket channel read loop, %s", c.self.Id)
	defer func() {
		logger.Debugf("websocket channel read close, %s", c.self.Id)
		c.Close()
	}()

	c.conn.SetReadLimit(c.config.ReadLimit)
	c.conn.SetReadDeadline(c.config.ReadDeadline())
	if c.isClient {
		c.conn.SetPongHandler(func(data string) error {
			logger.Warnf("websocket channel recv pong message, %s", data)
			c.conn.SetReadDeadline(c.config.ReadDeadline())
			return nil
		})
	} else {
		c.conn.SetPingHandler(func(data string) error {
			logger.Warnf("websocket channel recv ping message %s", data)
			c.conn.SetReadDeadline(c.config.ReadDeadline())
			if err := c.conn.WriteControl(websocket.PongMessage, []byte{}, c.config.WriteDeadline()); err != nil {
				logger.Errorf("websocket channel write pong message error, %s", err.Error())
				return err
			}
			return nil
		})
	}

	for {
		select {
		case <-c.doneChan:
			logger.Warnf("websocket channel done closed, %s", c.self.Id)
			return
		default:
			messageType, message, err := c.conn.ReadMessage()
			if err != nil {
				return
			}
			logger.Debugf("websocket channel read message<%d> %s", messageType, message)
			c.inChan <- message
		}
	}
}

// writeMessage -
func (c *WebsocketChannel) writeMessage(messageType int, data []byte) error {
	err := c.conn.WriteMessage(messageType, data)
	if err != nil {
		return err
	}
	c.conn.SetWriteDeadline(c.config.WriteDeadline())
	return nil
}

func (c *WebsocketChannel) SendSafe(data []byte) error {
	fmt.Println(">>> websocket channel send safe", string(data))
	select {
	case c.outChan <- data:
		return nil
	case <-c.doneChan:
		return channel.ErrDoneChannelClosed
	default:
		return channel.ErrOutChannelFull
	}
}

// WriteLoop -
func (c *WebsocketChannel) WriteLoop() {
	logger.Debugf("websocket channel write loop, %s", c.self.Id)
	defer func() {
		logger.Debugf("websocket channel write close, %s", c.self.Id)
		defer c.conn.Close()
	}()

	ticker := time.NewTicker(c.config.Keepalive())
	for {
		select {
		case <-c.doneChan:
			logger.Warnf("websocket channel done closed, %s", c.self.Id)
			return
		case message, ok := <-c.outChan:
			if !ok {
				logger.Warnf("websocket channel out chan closed, %s", c.self.Id)
				return
			}
			if err := c.writeMessage(websocket.BinaryMessage, message); err != nil {
				logger.Errorf("websocket channel write message error, %s", err.Error())
				return
			}
		case <-ticker.C:
			if c.isClient {
				if err := c.writeMessage(websocket.PingMessage, []byte{}); err != nil {
					logger.Errorf("websocket channel write ping message error, %s", err.Error())
					return
				}
				logger.Warnf("websocket channel send ping message")
			}
		}
	}
}

// Colse -
func (c *WebsocketChannel) Close() error {
	c.closeOnce.Do(func() {
		logger.Warnf("websocket channel close, %s", c.self.Id)
		c.conn.SetReadDeadline(time.Now())
		c.conn.SetWriteDeadline(time.Now())
		close(c.doneChan)
	})
	return nil
}
