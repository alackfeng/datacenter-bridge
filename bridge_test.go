package datacenterbridge_test

import (
	"context"
	"testing"
	"time"

	dcb "github.com/alackfeng/datacenter-bridge"
	"github.com/alackfeng/datacenter-bridge/channel"
)

func mockServerConfig() *dcb.Configure {
	config, _ := dcb.LoadConfigure("./config/config.yaml")
	return config
}

func TestDCenterBridge(t *testing.T) {
	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// bridge server mock.
	ds := dcb.NewDCenterBridge(ctx, done, mockServerConfig())
	if err := ds.ListenAndServe(); err != nil { //
		t.Error(err)
	}
	go ds.ChannelsLoop(func(ch channel.Channel, data []byte) {
		t.Log("get data:", string(data))
	}, func(ch channel.Channel) {})
	ds.WaitQuit()

	time.Sleep(time.Second * 3)

	// bridge client mock.
	dc := dcb.NewDCenterBridgeWithClient(ctx, done,
		dcb.AppInfo{Zone: "us", Service: "gw-dcb-service", Id: "xxx"},
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
