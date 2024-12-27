package datacenterbridge_test

import (
	"context"
	"testing"
	"time"

	dcb "github.com/alackfeng/datacenter-bridge"
	"github.com/alackfeng/datacenter-bridge/logger"
)

func mockServerConfig() *dcb.Configure {
	config := &dcb.Configure{
		Zone:    "us",
		Service: "gw-dcb-service",
		Id:      "main-us-1",
		Log:     *logger.NewLogConfigure(),
		Discovery: dcb.DiscoveryConfigure{
			Consul: dcb.ConsulConfigure{
				Up:   false,
				Host: "http://127.0.0.1:8500",
			},
			Etcd: dcb.EtcdConfigure{
				Up: true,
				Endpoints: []string{
					"127.0.0.1:23791",
				},
				Prefix:     "/dcbridge",
				GrantedTTL: 10,
			},
		},
		Servers: dcb.ServerConfigure{
			Ws: dcb.WebsocketConfigure{
				Up:   true,
				Host: "ws://127.0.0.1:9500/bridge",
			},
		},
	}
	return config
}

func TestDCenterBridge(t *testing.T) {
	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// bridge server mock.
	ds := dcb.NewDCenterBridgeWithServer(ctx, done, mockServerConfig())
	if err := ds.ListenAndServe(); err != nil { //
		t.Error(err)
	}
	go ds.ChannelsLoop(func(data []byte) {
		t.Log("get data:", string(data))
	})
	ds.WaitQuit()

	time.Sleep(time.Second * 3)

	// bridge client mock.
	dc := dcb.NewDCenterBridgeWithClient(ctx, done,
		dcb.NewSelf("us", "gw-dcb-service", "xxx"),
		dcb.ConsulConfigure{
			Up:   false,
			Host: "http://127.0.0.1:8500",
		},
		dcb.EtcdConfigure{
			Up: true,
			Endpoints: []string{
				"127.0.0.1:23791",
			},
			Prefix:     "/dcbridge",
			GrantedTTL: 10,
		},
	)
	if err := dc.SendData("us", "gw-dcb-service", []byte("hello")); err != nil {
		t.Error(err)
	}

	// wait quit all.
	<-done
}
