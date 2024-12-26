package websocket

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
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
func (s *WebsocketChannel) loginHeader() http.Header {
	header := http.Header{}
	selfJson, err := json.Marshal(s.self)
	if err != nil {
		return nil
	}
	header.Add(DcBridgeAuthHeader, base64.StdEncoding.EncodeToString(selfJson))
	return header
}

// Connect -
func (s *WebsocketClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: s.config.HandshakeDeadline(),
		ReadBufferSize:   s.config.ReadBufferSize,
		WriteBufferSize:  s.config.WriteBufferSize,
	}
	if s.config.Scheme == "wss" { // safe???.
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	conn, _, err := dialer.Dial(s.config.Url(), s.loginHeader())
	if err != nil {
		return err
	}
	s.init(conn)
	return nil
}
