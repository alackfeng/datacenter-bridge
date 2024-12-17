package discovery

import (
	"context"
	"encoding/json"
	"fmt"
)

// ConsulDiscovery - use consul discovery.
type ConsulDiscovery struct {
	baseUrl string
	HttpClient
}

// NewConsulDiscovery -
func NewConsulDiscovery(baseUrl string) *ConsulDiscovery {
	return &ConsulDiscovery{
		HttpClient: *NewHttpClient(),
		baseUrl:    baseUrl,
	}
}

// HealthServiceResp -
type HealthServiceResp struct {
}

// GetService - 获取某服务列表.
func (c *ConsulDiscovery) GetService(ctx context.Context, zone, serviceName string) ([]string, error) {
	res, err := c.Get(ctx, fmt.Sprintf("%s/v1/health/service/%s?dc=%s", c.baseUrl, serviceName, zone), nil)
	if err != nil {
		fmt.Println("consul discovery GetService err ", err)
		return nil, err
	}
	fmt.Println("consul discovery GetService ", string(res))
	var resp HealthServiceResp
	if err := json.Unmarshal(res, &resp); err != nil {
		return nil, err
	}
	return nil, nil
}
