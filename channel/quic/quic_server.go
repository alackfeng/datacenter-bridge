package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"

	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
	"github.com/quic-go/quic-go"
)

// QuicServer -
type QuicServer struct {
	self   discovery.Service
	config QuicConfig
}

// NewQuicServer -
func NewQuicServer(self *discovery.Service, config *QuicConfig) *QuicServer {
	return &QuicServer{
		self:   *self,
		config: *config,
	}
}

// generateTLSConfig -
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		logger.Errorf("quic server generate key err: %v", err)
		return nil
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		logger.Errorf("quic server create cert err: %v", err)
		return nil
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		logger.Errorf("quic server create cert err: %v", err)
		return nil
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-bridge"},
	}
}

// ListenAndServe -
func (s *QuicServer) ListenAndServe(ctx context.Context, channelChan chan<- channel.Channel) {
	listener, err := quic.ListenAddr(s.config.Addr(), generateTLSConfig(), &quic.Config{
		KeepAlivePeriod:      s.config.Keepalive(),
		MaxIdleTimeout:       s.config.MaxIdleTimeout(),
		HandshakeIdleTimeout: s.config.HandshakeIdleTimeout(),
	})
	if err != nil {
		logger.Errorf("quic server listen err: %v", err)
		return
	}
	defer listener.Close()
	logger.Debugf("start quic server <quic://%s>...", s.config.Host)
	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			logger.Errorf("quic server accept err: %v", err)
			select {
			case <-ctx.Done():
				return
			default:
				return
			}
		}
		go s.handleConnection(ctx, conn, channelChan)
	}
}

func (s *QuicServer) handleConnection(ctx context.Context, conn quic.Connection, channelChan chan<- channel.Channel) {
	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			logger.Errorf("quic server accept stream err: %v", err)
			return
		}
		logger.Infof("new quic stream from %s ", conn.RemoteAddr().String())
		// waiting login???
		var req QuicMessage
		if err := json.NewDecoder(stream).Decode(&req); err != nil {
			logger.Errorf("quic server accept stream login err: %v", err)
			return
		}
		var peer discovery.Service
		if err := req.Unpack(&peer); err != nil {
			logger.Errorf("quic server accept stream unpack err: %v", err)
			return
		}
		logger.Infof("new quic stream from %s data: %v.", conn.RemoteAddr().String(), peer)
		if err := json.NewEncoder(stream).Encode(Pack(TextMessage, &s.self)); err != nil {
			logger.Errorf("quic server accept stream login err: %v", err)
			return
		}
		quicChannel := newQuicServerChannel(&s.self, &peer, &s.config).init(conn, stream)
		channelChan <- quicChannel
		logger.Debugf("accept quic channel: %v", quicChannel)
	}
}
