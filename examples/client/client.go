package main

import (
	"fmt"

	datacenterbridge "github.com/alackfeng/datacenter-bridge"
)

func main() {
	fmt.Println("client::main - begin.")
	dcBridge := datacenterbridge.NewDCenterBridge()
	if err := dcBridge.Connect(""); err != nil {
		fmt.Println("client::main - connect error:", err)
		return
	}
	fmt.Println("client::main - end...")
}
