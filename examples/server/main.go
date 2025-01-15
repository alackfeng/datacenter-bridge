package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	dcb "github.com/alackfeng/datacenter-bridge"
	"github.com/alackfeng/datacenter-bridge/channel"
	"github.com/alackfeng/datacenter-bridge/logger"
)

var configFile string

// var quit bool = false
var dcBridge dcb.Datacenter

func consolePanel(ctx context.Context, cancel context.CancelFunc) {
	var help = `
>>> help: (quit|help|listen|channel_create|channel_delete|channel_list)
  - quit(q): 退出
  - help(h): 关闭
  - listen(l): 监听服务列表
  - channel_list(cl): 通道列表
  - channel_create(cc): 创建通道 (channel_create zone service id host)
	> eg: channel_create cn-001 gw-dcb-service server1 ws://127.0.0.1:9500/bridge
  - channel_send(cs): 发送数据 (channel_send zone service data)
	> eg: channel_send cn-001 gw-dcb-service "hello,world!"
  - channel_delete(cd): 删除通道 (channel_delete zone service id)
	> eg: channel_delete cn-001 gw-dcb-service
<<< over.
	`

	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("\ninput:> ")
			scanner.Scan()
			input := scanner.Text()
			if err := scanner.Err(); err != nil {
				fmt.Println("server::consolePanel - read input error: ", err)
				continue
			}
			options := strings.Split(input, " ")
			if len(options) == 0 || options[0] == "" {
				fmt.Println(help)
				continue
			}
			switch cmd := options[0]; strings.ToLower(cmd) {
			case "quit", "q", "qu", "qui":
				fmt.Println("server::consolePanel - quit...")
				cancel()
				return
			case "help", "h", "he", "hel":
				fmt.Println(help)
			case "listen", "l", "li", "lis":
				fmt.Printf("GetListenerList: \n%+v\n", dcBridge.GetListenerList())
			case "channel_list", "cl", "channellist":
				fmt.Printf("GetChannelList: \n%+v\n", dcBridge.GetChannelList())
			case "channel_create", "cc", "channelcreate":
				if len(options) < 4 {
					fmt.Println("!!!channel_create need 5 params: zone service id host!!!")
					continue
				}
				if _, err := dcBridge.CreateChannelForTest(options[1], options[2], options[3], options[4]); err != nil {
					fmt.Println("!!!channel_create error: ", err)
				} else {
					fmt.Println("channel_create ok.")
				}
			case "channel_delete", "cd", "channeldelete":
				if len(options) < 3 {
					fmt.Println("!!!channel_delete need 3 params: zone service!!!")
					continue
				}
				if err := dcBridge.DeleteChannel(options[1], options[2], ""); err != nil {
					fmt.Println("!!!channel_delete error: ", err)
				} else {
					fmt.Println("channel_delete ok.")
				}
			case "channel_send", "cs", "channelsend":
				if len(options) < 3 {
					fmt.Println("!!!channel_send need 3 params: zone service data!!!")
					continue
				}
				if err := dcBridge.SendData(options[1], options[2], []byte(options[3])); err != nil {
					fmt.Println("!!!channel_send error: ", err)
				} else {
					fmt.Println("channel_send ok.")
				}
			default:
				fmt.Println("!!!unknown command:>", input, "<!!!")
				fmt.Println(help)
			}
		}
	}
}

func main() {

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

	dcBridge = dcb.NewDCenterBridge(ctx, done, config)
	if err := dcBridge.ListenAndServe(); err != nil {
		logger.Fatalf("server::main - server error: %v", err)
	}
	go dcBridge.ChannelsLoop(func(ch channel.Channel, data []byte) {
		fmt.Println("server::main - get data:", string(data))
	}, func(ch channel.Channel) {
		fmt.Println("server::main - channel closed.")
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	time.Sleep(time.Second * 1)
	fmt.Println("server::main - running.")
	go consolePanel(ctx, cancel)

	select {
	case <-sig:
		fmt.Println("server::main - sig quit...")
		cancel()
	case <-ctx.Done():
		fmt.Println("server::main - cancel done...")
	}
	go dcBridge.WaitQuit()
	<-done
	fmt.Println("server::main - end...")
}
