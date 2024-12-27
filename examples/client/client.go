package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	datacenterbridge "github.com/alackfeng/datacenter-bridge"
	"github.com/alackfeng/datacenter-bridge/logger"
)

var mode string
var discoveryHost string

func main() {
	fmt.Println("client::main - begin.")

	flag.StringVar(&mode, "model", "debug", "run as release mode")
	flag.StringVar(&discoveryHost, "discoveryHost", "http://127.0.0.1:8500", "consul discovery host")
	flag.Parse()

	fmt.Printf("client::main - release mode<%v> \n", mode)
	logger.InitLogger(mode, logger.NewLogConfigure())

	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	dcBridge := datacenterbridge.NewDCenterBridgeWithClient(ctx, done,
		datacenterbridge.AppInfo{
			Zone:    "us",
			Service: "gw-dcb-service",
			Id:      "xxx",
		}, datacenterbridge.ConsulConfigure{
			Up:   true,
			Host: discoveryHost,
		},
	)
	go func() {
		dcBridge.ChannelsLoop(func(data []byte) {
			fmt.Println("client::main - get data:", string(data))
		})
	}()
	// time.Sleep(time.Second)

	ch, err := dcBridge.CreateChannel("us", "gw-dcb-service")
	if err != nil {
		fmt.Println("client::main - connect error:", err)
		os.Exit(1)
	}
	if err := ch.SendSafe([]byte("hello")); err != nil {
		fmt.Println("client::main - send error:", err)
	}

	fmt.Println("client::main - running.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	s := <-sig
	fmt.Println("client::main - sig quit...", s)
	cancel()
	fmt.Println("client::main - end...")
}
