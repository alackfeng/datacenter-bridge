package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	datacenterbridge "github.com/alackfeng/datacenter-bridge"
	"github.com/alackfeng/datacenter-bridge/logger"
)

var release bool
var host string
var consulHost string

func main() {
	flag.BoolVar(&release, "release", false, "run as release mode")
	flag.StringVar(&host, "host", "ws://10.16.3.66:9500/bridge", "server host")
	flag.StringVar(&consulHost, "consul", "http://127.0.0.1:8500", "consul host")
	flag.Parse()

	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())

	logger.InitLogger(release, logger.NewLogConfigure())

	fmt.Println("server::main - run as server, release mode", release)
	config := datacenterbridge.NewConfigure()
	config.Discovery.Consul = datacenterbridge.ConsulConfigure{
		Up:   true,
		Host: consulHost,
	}
	config.Servers.Ws = datacenterbridge.WebsocketConfigure{
		Up:   true,
		Host: host,
	}
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
