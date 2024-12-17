package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alackfeng/datacenter-bridge/logger"
	"github.com/gorilla/websocket"
)

var allowedOrigins = []string{"*"}

// WebsocketServer -
type WebsocketServer struct {
	config   WebsocketConfig
	upgrader websocket.Upgrader
	// client   sync.Map
}

// NewWebsocketServer -
func NewWebsocketServer(config *WebsocketConfig) *WebsocketServer {
	return &WebsocketServer{
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

// acceptWebsocket -
func (s *WebsocketServer) acceptWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	wsChannel := newWebsocketServerChannel(&s.config).init(conn)
	fmt.Println("accept websocket channel: ", wsChannel)
}

// ListenAndServe -
func (s *WebsocketServer) ListenAndServe(ctx context.Context) {
	server := http.Server{
		Addr: s.config.Host(),
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.Debugf("websocket health ok. %s", r.RemoteAddr)
		w.Write([]byte("ok"))
	})
	http.HandleFunc(s.config.Prefix, s.acceptWebsocket)

	go func() {
		logger.Debugf("start websocket server <http://%s>...", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			logger.Errorf("websocket server error: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Error("websocket server shutdown...")

	stop, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(stop); err != nil {
		logger.Errorf("websocket server shutdown error: %v", err)
	}
	logger.Error("websocket server gracefully stopped")
}
