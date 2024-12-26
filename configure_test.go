package datacenterbridge_test

import (
	"testing"

	dcb "github.com/alackfeng/datacenter-bridge"
	"github.com/alackfeng/datacenter-bridge/logger"
)

func TestConfigure(t *testing.T) {

	config := &dcb.Configure{
		Zone:    "us-001",
		Service: "gw-dcb-service",
		Id:      "s001",
		Log: logger.LogConfigure{
			FilePath:    "./logs/app.log",
			FileMaxSize: 28,
			FileMaxAge:  100,
		},
		Discovery: dcb.DiscoveryConfigure{
			Consul: dcb.ConsulConfigure{
				Up:   true,
				Host: "http://127.0.0.1:8500",
			},
			Etcd: dcb.EtcdConfigure{
				Up:         false,
				Endpoints:  []string{"127.0.0.1:2379"},
				Prefix:     "/dcbridge",
				GrantedTTL: 10,
			},
		},
		Servers: dcb.ServerConfigure{
			Ws: dcb.WebsocketConfigure{
				Up:   true,
				Host: "ws://127.0.0.1:9500/bridge",
			},
			Wss: dcb.WebsocketsConfigure{
				Up:       true,
				Host:     "wss://127.0.0.1:9500/bridge",
				CertFile: "./testdata/cacert.pem",
				KeyFile:  "./testdata/cakey.pem",
			},
			Quic: dcb.QuicConfigure{
				Up:   true,
				Host: "quic://127.0.0.1:9501",
			},
		},
	}
	t.Log(config.Self().Id == config.Id)
}
