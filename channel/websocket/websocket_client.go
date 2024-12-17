package websocket

import (
	"github.com/gorilla/websocket"
)

// WebsocketClient -
type WebsocketClient struct {
	*WebsocketChannel
}

// NewWebsocketClient -
func NewWebsocketClient(config *WebsocketConfig) *WebsocketClient {
	return &WebsocketClient{
		WebsocketChannel: newWebsocketClientChannel(config),
	}
}

// Connect -
func (s *WebsocketClient) Connect() error {
	dialer := websocket.Dialer{
		HandshakeTimeout: s.config.HandshakeDeadline(),
		ReadBufferSize:   s.config.ReadBufferSize,
		WriteBufferSize:  s.config.WriteBufferSize,
	}
	conn, _, err := dialer.Dial(s.config.Url(), nil)
	if err != nil {
		return err
	}
	s.init(conn)
	return nil
}
