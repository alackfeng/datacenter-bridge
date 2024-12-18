package websocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/gorilla/websocket"
)

const DcBridgeAuthHeader string = "Bridge"

// WebsocketClient -
type WebsocketClient struct {
	*WebsocketChannel
}

// NewWebsocketClient -
func NewWebsocketClient(self *discovery.Service, peer *discovery.Service) *WebsocketClient {
	return &WebsocketClient{
		WebsocketChannel: newWebsocketClientChannel(self, peer),
	}
}

// LoginHeader - 登录认证请求头.
func (s *WebsocketChannel) LoginHeader() http.Header {
	header := http.Header{}
	selfJson, err := json.Marshal(s.self)
	if err != nil {
		return nil
	}
	header.Add(DcBridgeAuthHeader, base64.StdEncoding.EncodeToString(selfJson))
	return header
}

// Connect -
func (s *WebsocketClient) Connect() error {
	dialer := websocket.Dialer{
		HandshakeTimeout: s.config.HandshakeDeadline(),
		ReadBufferSize:   s.config.ReadBufferSize,
		WriteBufferSize:  s.config.WriteBufferSize,
	}
	fmt.Println("websocket connect .")
	conn, _, err := dialer.Dial(s.config.Url(), s.LoginHeader())
	if err != nil {
		return err
	}
	fmt.Println("websocket connect 1.")
	s.init(conn)
	fmt.Println("websocket connect 2.")
	return nil
}
