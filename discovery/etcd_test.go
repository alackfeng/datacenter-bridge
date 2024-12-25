package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/alackfeng/datacenter-bridge/discovery"
)

func TestEtcdRegistry(t *testing.T) {

	ctx := context.Background()
	dis := discovery.NewEtcdRegistry([]string{"127.0.0.1:23791"}, "/dcbridge", 10)
	if dis == nil {
		t.Error("registry is nil")
	}
	service := discovery.Service{
		Zone:    "us",
		Id:      "main-us-1",
		Service: "gw-dcb-service",
		Host:    "ws://127.0.0.1:7900/ws",
		Tag:     "primary",
	}
	if err := dis.Register(ctx, service); err != nil {
		t.Error("regsiter error", err)
	}
	services, err := dis.GetServices(ctx, service.Zone, service.Id)
	if err != nil {
		t.Error("get service error", err)
	}
	t.Log(services)

	dis.Watch(ctx)

	t.Log("watching...")
	stop, cancel := context.WithTimeout(ctx, time.Second*30)
	<-stop.Done()
	t.Log("quit...")
	cancel()
}
