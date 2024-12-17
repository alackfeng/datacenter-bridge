package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	datacenterbridge "github.com/alackfeng/datacenter-bridge"
)

var release bool
var discoveryHost string

func main() {
	fmt.Println("client::main - begin.")

	flag.BoolVar(&release, "release", false, "run as release mode")
	flag.StringVar(&discoveryHost, "discoveryHost", "http://127.0.0.1:8500", "consul discovery host")
	flag.Parse()

	fmt.Printf("client::main - release mode<%v> \n", release)

	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	dcBridge := datacenterbridge.NewDCenterBridge(ctx, done, discoveryHost)
	if err := dcBridge.CreateChannel("us", "gw-dcb-service"); err != nil {
		fmt.Println("client::main - connect error:", err)
	}

	fmt.Println("client::main - running.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	s := <-sig
	fmt.Println("client::main - sig quit...", s)
	cancel()
	fmt.Println("client::main - end...")
}
