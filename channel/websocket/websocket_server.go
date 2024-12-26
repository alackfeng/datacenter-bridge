package websocket

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
	"github.com/gorilla/websocket"
)

var allowedOrigins = []string{"*"}

// WebsocketServer -
type WebsocketServer struct {
	self     discovery.Service
	config   WebsocketConfig
	upgrader websocket.Upgrader
}

// NewWebsocketServer -
func NewWebsocketServer(self *discovery.Service, config *WebsocketConfig) *WebsocketServer {
	return &WebsocketServer{
		self:   *self,
		config: *config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, allowed := range allowedOrigins {
					if allowed == "*" || origin == allowed {
						return true
					}
				}
				return true
			},
			ReadBufferSize:   config.ReadBufferSize,
			WriteBufferSize:  config.WriteBufferSize,
			HandshakeTimeout: config.HandshakeDeadline(),
		},
	}
}

// acceptWebsocket - create ws channel.
func (s *WebsocketServer) acceptWebsocket(channelChan chan<- channel.Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 头认证.
		req := r.Header.Get(DcBridgeAuthHeader)
		if req == "" {
			logger.Errorf("websocket auth error: no http header: Bridge")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		reqBody, err := base64.StdEncoding.DecodeString(req)
		if err != nil {
			logger.Errorf("websocket auth error: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var res discovery.Service
		if err := json.Unmarshal(reqBody, &res); err != nil {
			logger.Errorf("websocket auth error: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// 升级ws.
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		wsChannel := newWebsocketServerChannel(&s.self, &res, &s.config).init(conn)
		channelChan <- wsChannel
		logger.Debugf("accept websocket channel: %v", wsChannel)
	}
}

// handleHealth - consul request health.
func (s *WebsocketServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("websocket health ok. %s", r.RemoteAddr)
	w.Write([]byte("ok"))
}

func (s *WebsocketServer) safe() bool {
	return s.config.Scheme == "wss"
}

// ListenAndServe -
func (s *WebsocketServer) ListenAndServe(ctx context.Context, channelChan chan<- channel.Channel) {
	isSafe := s.safe()

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    s.config.Host(),
		Handler: mux,
	}
	if isSafe {
		server.TLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
		}
	}
	mux.HandleFunc("/health", s.handleHealth)                       // prefix: /health .
	mux.HandleFunc(s.config.Prefix, s.acceptWebsocket(channelChan)) // prefix: /bridge .

	go func() {
		logger.Debugf("start websocket server <http://%s>...", server.Addr)
		if isSafe {
			if err := server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile); err != nil {
				logger.Errorf("websockets server error: %v", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil {
				logger.Errorf("websocket server error: %v", err)
			}
		}
	}()

	<-ctx.Done()
	logger.Info("websocket server shutdown...")

	stop, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(stop); err != nil {
		logger.Errorf("websocket server shutdown error: %v", err)
	}
	logger.Info("websocket server gracefully stopped")
}
