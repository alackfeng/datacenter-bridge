package quic

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/discovery"
	"github.com/alackfeng/datacenter-bridge/logger"
	"github.com/quic-go/quic-go"
)

// QuicChannel -
type QuicChannel struct {
	self        discovery.Service // 身份标识.
	peer        discovery.Service // 对端身份.
	config      QuicConfig
	conn        quic.Connection
	stream      quic.Stream
	closeOnce   sync.Once
	doneChan    chan struct{}
	inChan      chan []byte
	outChan     chan []byte
	isConnected bool
	isClient    bool
}

var _ channel.Channel = (*QuicChannel)(nil)

// newQuicServerChannel -
func newQuicServerChannel(self *discovery.Service, peer *discovery.Service, config *QuicConfig) *QuicChannel {
	return newQuicChannel(self, peer, config, false)
}

// newQuicClientChannel -
func newQuicClientChannel(self *discovery.Service, peer *discovery.Service) *QuicChannel {
	return newQuicChannel(self, peer, NewQuicConfig(peer.Host), true)
}

// newQuicChannel -
func newQuicChannel(self *discovery.Service, peer *discovery.Service, config *QuicConfig, isClient bool) *QuicChannel {
	return &QuicChannel{
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

// Close implements channel.Channel.
func (q *QuicChannel) Close() error {
	q.closeOnce.Do(func() {
		logger.Warnf("quic channel close, %s", q.self.Id)
		close(q.doneChan)
	})
	return nil
}

// DoneChan implements channel.Channel.
func (q *QuicChannel) DoneChan() chan struct{} {
	return q.doneChan
}

// InChan implements channel.Channel.
func (q *QuicChannel) InChan() chan []byte {
	return q.inChan
}

func (q *QuicChannel) init(conn quic.Connection, stream quic.Stream) *QuicChannel {
	q.conn = conn
	q.stream = stream
	q.isConnected = true
	q.self.Host = q.conn.LocalAddr().String()
	return q
}

// ID implements channel.Channel.
func (q *QuicChannel) ID() string {
	return q.peer.Id
}

// Key implements channel.Channel.
func (q *QuicChannel) Key() string {
	return q.peer.Key()
}

func (c *QuicChannel) Self() discovery.Service {
	return c.self
}
func (c *QuicChannel) Peer() discovery.Service {
	return c.peer
}
func (c *QuicChannel) Info() channel.ChannelInfo {
	return channel.ChannelInfo{
		Local: c.self,
		Peer:  c.peer,
	}
}

// String -
func (c *QuicChannel) String() string {
	return fmt.Sprintf("[%s:%s:%s]%s=>[%s:%s:%s]%s",
		c.self.Service, c.self.Zone, c.self.Id, c.self.Host,
		c.peer.Service, c.peer.Zone, c.peer.Id, c.peer.Host)
}

// ReadLoop implements channel.Channel.
func (q *QuicChannel) ReadLoop() {
	defer func() {
		q.Close()
	}()

	decode := json.NewDecoder(q.stream)
	for {
		select {
		case <-q.doneChan:
			logger.Warnf("quic channel done closed, %s", q.self.Id)
			return
		default:
			var data QuicMessage
			if err := decode.Decode(&data); err != nil {
				logger.Errorf("read data error: %v", err)
				return
			} else if data.Type == PingMessage {
				logger.Warnf("quic channel receive message type: %d", data.Type)
				if err := q.writeMessage(PongMessage, []byte{}); err != nil {
					logger.Errorf("quic channel write message error, %s", err.Error())
					return
				}
				continue
			} else if data.Type == PongMessage {
				logger.Warnf("quic channel receive message type: %d", data.Type)
				continue
			}
			fmt.Println("quic channel read data: ", data.Timestamp, string(data.Payload))
			q.inChan <- data.Payload
		}
	}
}

// writeMessage -
func (q *QuicChannel) writeMessage(messageType uint8, data []byte) error {
	logger.Debugf("quic channel write message<%d>: %d, %s", messageType, len(data), q.self.Id)
	if msg, err := json.Marshal(NewQuicMessage(messageType, data)); err != nil {
		return err
	} else if n, err := q.stream.Write(msg); err != nil {
		return err
	} else if n != len(msg) {
		fmt.Println("n, len ", n, len(msg))
		return channel.ErrWriteNotFull
	}
	return nil
}

// SendSafe implements channel.Channel.
func (q *QuicChannel) SendSafe(data []byte) error {
	select {
	case q.outChan <- data:
		return nil
	case <-q.doneChan:
		return channel.ErrDoneChannelClosed
	default:
		return channel.ErrOutChannelFull
	}
}

// WriteLoop implements channel.Channel.
func (q *QuicChannel) WriteLoop() {
	defer func() {
		q.stream.Close()
		q.conn.CloseWithError(0, "")
	}()

	ticker := time.NewTicker(q.config.Keepalive())
	for {
		select {
		case <-q.doneChan:
			logger.Warnf("quic channel done closed, %s", q.self.Id)
			return
		case data, ok := <-q.outChan:
			if !ok {
				logger.Warnf("quic channel out chan closed, %s", q.self.Id)
				return
			}
			if err := q.writeMessage(BinaryMessage, data); err != nil {
				logger.Errorf("quic channel write message error, %s", err.Error())
				return
			}
		case <-ticker.C:
			if q.isClient {
				if err := q.writeMessage(PingMessage, []byte{}); err != nil {
					logger.Errorf("quic channel write ping message error, %s", err.Error())
					return
				}
				logger.Warnf("quic channel send ping message")
			}
		}
	}
}
