package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/channel/quic"
	"github.com/alackfeng/datacenter-bridge/channel/websocket"
	"github.com/alackfeng/datacenter-bridge/discovery"
)

func TestWebsocketChannel(t *testing.T) {
	url := "ws://10.16.3.206:9500/bridge"
	self1 := &discovery.Service{
		Zone:    "us-001",
		Service: "gw-dcb-service",
		Id:      "s001",
		Host:    url,
		Tag:     "primary",
	}
	self2 := &discovery.Service{
		Zone:    "us-001",
		Service: "gw-dcb-service",
		Id:      "s002",
	}

	ctx, cancel := context.WithCancel(context.Background())
	// mock websocket server.
	chs := make(chan channel.Channel)
	var sch struct {
		ch channel.Channel
		in chan []byte
	}
	go func() {
		s := websocket.NewWebsocketServer(self1, websocket.NewWebsocketConfig(url))
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case ch := <-chs:
					go ch.ReadLoop()
					go ch.WriteLoop()
					sch.ch = ch
					sch.in = ch.InChan()
				case data, ok := <-sch.in:
					if !ok {
						continue
					}
					t.Log("server recv: ", string(data))
					if err := sch.ch.SendSafe(data); err != nil {
						t.Error("server send error", err)
					}
				}
			}
		}()
		s.ListenAndServe(ctx, chs)
	}()

	time.Sleep(time.Second * 3)
	// mock websocket client.
	c := websocket.NewWebsocketClient(self2, self1)
	if err := c.Connect(ctx); err != nil {
		t.Error("websocket client connect error", err)
	}
	go c.ReadLoop()
	go c.WriteLoop()
	time.Sleep(time.Second)
	if err := c.SendSafe([]byte("hello")); err != nil {
		t.Error("websocket client send error", err)
	}

	t.Log("waiting...")
	select {
	case <-ctx.Done():
		t.Log("timeout quit")
		cancel()
	case data, ok := <-c.InChan():
		if !ok {
			t.Log("channel closed")
		}
		t.Log("client recv: ", string(data))
		cancel()
	}
	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, os.Interrupt)
	// stop, stopCancel := context.WithTimeout(ctx, time.Second*100)
	// select {
	// case <-stop.Done():
	// 	t.Log("timeout quit")
	// 	stopCancel()
	// 	cancel()
	// case <-sig:
	// 	t.Log("sig quit")
	// 	stopCancel()
	// 	cancel()
	// }
}

func TestQuicChannel(t *testing.T) {
	url := "quic://10.16.3.206:9500"
	self1 := &discovery.Service{
		Zone:    "us-001",
		Service: "gw-dcb-service",
		Id:      "s001",
		Host:    url,
		Tag:     "primary",
	}
	self2 := &discovery.Service{
		Zone:    "us-001",
		Service: "gw-dcb-service",
		Id:      "s002",
	}

	ctx, cancel := context.WithCancel(context.Background())
	// mock websocket server.
	chs := make(chan channel.Channel)
	var sch struct {
		ch channel.Channel
		in chan []byte
	}
	go func() {
		s := quic.NewQuicServer(self1, quic.NewQuicConfig(url))
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case ch := <-chs:
					go ch.ReadLoop()
					go ch.WriteLoop()
					sch.ch = ch
					sch.in = ch.InChan()
				case data, ok := <-sch.in:
					if !ok {
						continue
					}
					t.Log("quic server recv: ", string(data))
					if err := sch.ch.SendSafe(data); err != nil {
						t.Error("quic server send error", err)
					}
				}
			}
		}()
		s.ListenAndServe(ctx, chs)
	}()

	time.Sleep(time.Second * 3)
	// mock websocket client.
	c := quic.NewQuicClient(self2, self1)
	if err := c.Connect(ctx); err != nil {
		t.Error("quic client connect error", err)
	}
	go c.ReadLoop()
	go c.WriteLoop()
	time.Sleep(time.Second)
	if err := c.SendSafe([]byte("hello")); err != nil {
		t.Error("quic client send error", err)
	}

	t.Log("waiting...")
	select {
	case <-ctx.Done():
		t.Log("timeout quit")
		cancel()
	case data, ok := <-c.InChan():
		if !ok {
			t.Log("channel closed")
		}
		t.Log("client recv: ", string(data))
		cancel()
	}
}
