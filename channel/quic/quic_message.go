package quic

import (
	"encoding/json"
	"time"

	"github.com/alackfeng/datacenter-bridge/discovery"
)

var (
	TextMessage   uint8 = 1
	BinaryMessage uint8 = 2
	CloseMessage  uint8 = 8
	PingMessage   uint8 = 9
	PongMessage   uint8 = 10
)

const quicMessageVersion = "1"

// QuicMessage -
type QuicMessage struct {
	Timestamp int64  `json:"ts" comment:"时间戳ms"`
	Version   string `json:"v" comment:"版本号"`
	Type      uint8  `json:"t" comment:"协议类型"`
	Payload   []byte `json:"p" comment:"数据"`
}

// NewQuicMessage -
func NewQuicMessage(messageType uint8, payload []byte) *QuicMessage {
	return &QuicMessage{
		Version:   quicMessageVersion,
		Timestamp: time.Now().UnixMilli(),
		Type:      messageType,
		Payload:   payload,
	}
}

// Pack -
func Pack(messageType uint8, data interface{}) *QuicMessage {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return &QuicMessage{
		Version:   quicMessageVersion,
		Timestamp: time.Now().UnixMilli(),
		Type:      TextMessage,
		Payload:   payload,
	}
}

// Unpack -
func (q *QuicMessage) Unpack(v interface{}) error {
	err := json.Unmarshal(q.Payload, v)
	if err != nil {
		return err
	}
	return nil
}

// // pingMessage -
// func pingMessage() *QuicMessage {
// 	return &QuicMessage{
// 		Version:   quicMessageVersion,
// 		Timestamp: time.Now().UnixMilli(),
// 		Type:      PingMessage,
// 		Payload:   []byte{},
// 	}
// }

// loginInfo - 登录请求信息.
func loginInfo(req *discovery.Service) *QuicMessage {
	selfJson, err := json.Marshal(req)
	if err != nil {
		return nil
	}
	return &QuicMessage{
		Version:   quicMessageVersion,
		Timestamp: time.Now().UnixMilli(),
		Type:      TextMessage,
		Payload:   selfJson,
	}
}
