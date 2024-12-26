package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	datacenterbridge "github.com/alackfeng/datacenter-bridge"
	"github.com/alackfeng/datacenter-bridge/logger"
)

var release bool
var host string
var quicHost string
var consulHost string

func bridgeConfig() *datacenterbridge.Configure {
	config := datacenterbridge.NewConfigure()
	config.Zone = "us"
	config.Id = "main-us-1"
	config.Service = "gw-dcb-service"
	config.Discovery.Consul = datacenterbridge.ConsulConfigure{
		Up:   true,
		Host: consulHost,
	}
	config.Servers.Ws = datacenterbridge.WebsocketConfigure{
		Up:   true,
		Host: host,
	}
	config.Servers.Wss = datacenterbridge.WebsocketsConfigure{
		Up:       true,
		Host:     host,
		CertFile: "./certs/server/server.crt",
		KeyFile:  "./certs/server/server.key",
	}
	u, err := url.Parse(quicHost)
	if err != nil {
		logger.Fatalf("server::main - quic host parse error: %v", err)
	}
	fmt.Println("server::main - quic host:", u.Scheme, u.Host)
	config.Servers.Quic = datacenterbridge.QuicConfigure{
		Up:   true,
		Host: u.Host,
	}
	return config
}

func main() {
	flag.BoolVar(&release, "release", false, "run as release mode")
	flag.StringVar(&host, "host", "ws://10.16.3.66:9500/bridge", "quic server host")
	flag.StringVar(&quicHost, "quic", "quic://10.16.3.66:9501", "quic server host")
	flag.StringVar(&consulHost, "consul", "http://127.0.0.1:8500", "consul host")
	flag.Parse()

	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())

	logger.InitLogger(release, logger.NewLogConfigure())

	fmt.Println("server::main - run as server, release mode", release)
	config := bridgeConfig()
	dcBridge := datacenterbridge.NewDCenterBridgeWithConfig(ctx, done, config)
	if err := dcBridge.ListenAndServe(); err != nil {
		logger.Fatalf("server::main - server error: %v", err)
	}
	dcBridge.WaitQuit()
	fmt.Println("server::main - running.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	s := <-sig
	fmt.Println("server::main - sig quit...>", s)
	cancel()

	quit := false
	stop, stopCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer stopCancel()
	for {
		select {
		case <-done:
			fmt.Println("server::main - done...")
			quit = true
		case <-stop.Done():
			fmt.Println("server::main - stop timeout...")
			quit = true
		}
		if quit {
			break
		}
	}
	fmt.Println("server::main - end...")
}
