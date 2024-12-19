package quic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
	"github.com/quic-go/quic-go"
)

// QuicClient -
type QuicClient struct {
	*QuicChannel
}

// NewQuicClient -
func NewQuicClient(self *discovery.Service, peer *discovery.Service) *QuicClient {
	return &QuicClient{
		QuicChannel: newQuicClientChannel(self, peer),
	}
}

func (q *QuicClient) Connect(ctx context.Context) error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-bridge"},
	}
	fmt.Println(">>>>quic host: ", q.peer.Address())
	conn, err := quic.DialAddr(ctx, q.peer.Address(), tlsConf, nil)
	if err != nil {
		return err
	}
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return err
	}
	// 发送登录信息交换身份并验证.
	if err := json.NewEncoder(stream).Encode(loginInfo(&q.self)); err != nil {
		return err
	}
	var resp QuicMessage
	if err := json.NewDecoder(stream).Decode(&resp); err != nil {
		return err
	} else if resp.Type != TextMessage {
		return channel.ErrMessageTypeNotMatch
	}
	var peer discovery.Service
	if err := resp.Unpack(&peer); err != nil {
		return err
	}
	logger.Debugf("quic client connected to %v", peer)
	q.init(conn, stream)
	return nil
}
