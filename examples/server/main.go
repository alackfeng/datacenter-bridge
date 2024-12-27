package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	dcb "github.com/alackfeng/datacenter-bridge"
	"github.com/alackfeng/datacenter-bridge/logger"
)

// var release bool
// var host string
// var quicHost string
// var consulHost string
var configFile string

func main() {
	// flag.BoolVar(&release, "release", false, "run as release mode")
	// flag.StringVar(&host, "host", "ws://10.16.3.66:9500/bridge", "quic server host")
	// flag.StringVar(&quicHost, "quic", "quic://10.16.3.66:9501", "quic server host")
	// flag.StringVar(&consulHost, "consul", "http://127.0.0.1:8500", "consul host")
	flag.StringVar(&configFile, "config", "./examples/server/dcb_server1.config.yaml", "config file path")
	flag.Parse()

	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())

	config, err := dcb.LoadConfigure(configFile)
	if err != nil {
		logger.Fatalf("server::main - load config error: %v", err)
	}
	fmt.Printf("server::main - run as server, mode:%v.\n", config.Mode)
	fmt.Printf("server::main - config %s file: >>>>\n %+v <<<<.\n", configFile, config)
	logger.InitLogger(config.Mode, &config.Log)

	dcBridge := dcb.NewDCenterBridge(ctx, done, config)
	if err := dcBridge.ListenAndServe(); err != nil {
		logger.Fatalf("server::main - server error: %v", err)
	}
	dcBridge.ChannelsLoop(func(data []byte) {
		fmt.Println("server::main - get data:", string(data))
	})
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
