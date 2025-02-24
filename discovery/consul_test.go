package discovery_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/alackfeng/datacenter-bridge/discovery"
)

func TestConsulRegsitery(t *testing.T) {
	testcases := []struct {
		service string
		zone    string
		name    string
		code    int    // 期望值.
		remark  string // 说明.
	}{
		{service: "gw-dcb-service", zone: "us-001", name: "s001", code: 0, remark: "网关服务us区域"},
		{service: "gw-dcb-service", zone: "us-001", name: "s002", code: 0, remark: "网关服务us区域"},
		{service: "gw-dcb-service", zone: "cn-001", name: "s003", code: 0, remark: "网关服务cn区域"},
	}

	var diss []discovery.Discovery
	ctx := context.Background()
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.remark), func(t *testing.T) {
			dis := discovery.NewConsulRegistry("http://127.0.0.1:23791", "")
			if dis == nil {
				t.Error("registry is nil")
			}
			diss = append(diss, dis)
		})
	}
	for i := 0; i < 2; i++ {
		zone := "us-001"
		serivce := "gw-dcb-service"
		if i == 1 {
			zone = "cn-001"
			serivce = "gw-dcb-service"
		}
		services, err := diss[i].GetServices(ctx, zone, serivce)
		if err != nil {
			t.Error("get service error", err)
		}
		if i == 0 && len(services) != 2 {
			t.Error("service count is not 2")
		} else if i == 1 && len(services) != 1 {
			t.Error("service count is not 1")
			// } else {
			// 	t.Error("service is empty")
		}
		// t.Log(i, len(services), services)
	}
}
