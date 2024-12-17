package websocket

import (
	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/gorilla/websocket"
)

// WebsocketChannel - websocket通道对象, 进行收发操作.
type WebsocketChannel struct {
	config      WebsocketConfig
	conn        *websocket.Conn
	done        chan struct{}
	inChan      chan []byte
	outChan     chan []byte
	isConnected bool
	isClinet    bool
}

var _ channel.Channel = (*WebsocketChannel)(nil)

// newWebsocketServerChannel -
func newWebsocketServerChannel(config *WebsocketConfig) *WebsocketChannel {
	w := newWebsocketChannel(config, false)
	return w
}

// newWebsocketClientChannel -
func newWebsocketClientChannel(config *WebsocketConfig) *WebsocketChannel {
	return newWebsocketChannel(config, true)
}

// newWebsocketChannel -
func newWebsocketChannel(config *WebsocketConfig, isClinet bool) *WebsocketChannel {
	return &WebsocketChannel{
		config:      *config,
		isClinet:    isClinet,
		isConnected: false,
		done:        make(chan struct{}),
		inChan:      make(chan []byte, config.InChanCount),
		outChan:     make(chan []byte, config.OutChanCount),
	}
}

func (c *WebsocketChannel) init(conn *websocket.Conn) *WebsocketChannel {
	c.conn = conn
	c.isConnected = true
	return c
}

// Read -
func (c *WebsocketChannel) Read() {
	defer func() {
		close(c.done)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(c.config.ReadLimit)
	c.conn.SetReadDeadline(c.config.ReadDeadline())
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(c.config.ReadDeadline())
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		c.inChan <- message
	}
}

func (c *WebsocketChannel) Write() {
	for {
		select {
		case message := <-c.outChan:
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *WebsocketChannel) Colse() error {
	return nil
}
