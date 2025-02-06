package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/alackfeng/datacenter-bridge/channel/quic"
	"github.com/alackfeng/datacenter-bridge/channel/websocket"
	"github.com/alackfeng/datacenter-bridge/discovery"
)

var zone string
var id string
var name string
var host string

func hostPrefix(host string) string {
	if u, err := url.Parse(host); err == nil {
		return u.Scheme
	}
	return ""
}

func checkQuic() error {
	qc := quic.NewQuicClient(&discovery.Service{
		Zone:    zone,
		Service: name,
		Id:      "health-check-client-001",
	}, &discovery.Service{
		Zone:    zone,
		Service: name,
		Id:      id,
		Host:    host,
		Tag:     "primary",
	})
	return qc.Connect(context.Background())
}

func checkWebsocket() error {
	wc := websocket.NewWebsocketClient(&discovery.Service{
		Zone:    zone,
		Service: name,
		Id:      "health-check-client-001",
	}, &discovery.Service{
		Zone:    zone,
		Service: name,
		Id:      id,
		Host:    host,
		Tag:     "primary",
	})
	return wc.Connect(context.Background())
}

func main() {
	flag.StringVar(&zone, "zone", "cn-001", "zone")
	flag.StringVar(&id, "id", "health-check-client-001", "service id")
	flag.StringVar(&name, "name", "gw-dcb-service", "service name")
	// flag.StringVar(&host, "host", "quic://10.16.3.206:9500", "quic server url")
	flag.StringVar(&host, "host", "ws://10.16.3.66:9500/bridge", "quic server url")

	flag.Parse()

	fmt.Println("health check begin.", host)

	var err error
	switch scheme := hostPrefix(host); scheme {
	case "quic":
		err = checkQuic()
	case "ws", "wss":
		err = checkWebsocket()
	default:
		fmt.Println("health check not support scheme: ", scheme)
		os.Exit(2)
	}

	if err != nil {
		fmt.Println("health check error: ", err)
		os.Exit(2)
	} else {
		fmt.Println("health check success.")
		os.Exit(0)
	}
}
