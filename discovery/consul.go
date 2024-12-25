package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/alackfeng/datacenter-bridge/logger"
	httpUtil "github.com/alackfeng/datacenter-bridge/utils/http"
)

// ConsulDiscovery - use consul discovery.
type ConsulDiscovery struct {
	baseUrl string
	httpUtil.HttpClient
}

var _ Discovery = (*ConsulDiscovery)(nil)

// NewConsulDiscovery -
func NewConsulDiscovery(baseUrl string) Discovery {
	return &ConsulDiscovery{
		HttpClient: *httpUtil.NewHttpClient(),
		baseUrl:    baseUrl,
	}
}

// ConsulNode -
type ConsulNode struct {
	Id         string `json:"ID" comment:"consul node id"`
	Node       string `json:"Node" comment:"consul node name"`
	Datacenter string `json:"Datacenter" comment:"consul node zone"`
	Address    string `json:"Address" comment:"consul node address"`
}

// ConsulService -
type ConsulService struct {
	Id      string   `json:"ID" comment:"consul service id"`
	Service string   `json:"Service" comment:"consul service name"`
	Tags    []string `json:"Tags" comment:"consul service tags"`
	Address string   `json:"Address" comment:"consul service address"`
	Port    int      `json:"Port" comment:"consul service port"`
}

// ConsulCheck -
type ConsulCheck struct {
	Node        string `json:"Node" comment:"consul node name"`
	CheckID     string `json:"CheckID" comment:"consul check id"`
	Name        string `json:"Name" comment:"consul check name"`
	Status      string `json:"Status" comment:"consul check status"`
	Notes       string `json:"Notes" comment:"consul check notes"`
	Output      string `json:"Output" comment:"consul check output"`
	ServiceID   string `json:"ServiceID" comment:"consul check service id"`
	ServiceName string `json:"ServiceName" comment:"consul check service name"`
}

// HealthService -
type HealthService struct {
	Node    ConsulNode    `json:"Node" comment:"consul node"`
	Service ConsulService `json:"Service" comment:"consul service"`
	Checks  []ConsulCheck `json:"Checks" comment:"consul check"`
}

// To -
func (c HealthService) To() *Service {
	for _, check := range c.Checks {
		if check.Status != "passing" {
			return nil
		}
	}
	var host, keyword string
	for _, tag := range c.Service.Tags {
		fmt.Println("tag: ", tag)
		if strings.HasPrefix(tag, "key:") {
			keyword = tag
			continue
		}
		u, err := url.Parse(tag)
		if err != nil {
			fmt.Println("err: ", err)
			continue
		}
		host = u.String()
	}
	fmt.Println(">>", host, keyword, ".")
	return &Service{
		Zone:    c.Node.Datacenter,
		Service: c.Service.Service,
		Id:      c.Service.Id,
		Host:    host,
		Tag:     keyword,
	}
}

// GetServices - 获取某服务列表.
func (c *ConsulDiscovery) GetServices(ctx context.Context, zone, serviceName string) ([]Service, error) {
	res, err := c.Get(ctx, fmt.Sprintf("%s/v1/health/service/%s?dc=%s&passing", c.baseUrl, serviceName, zone), nil)
	if err != nil {
		logger.Errorf("consul discovery GetServices err: %v", err)
		return nil, err
	}
	fmt.Println("consul discovery GetServices ", string(res))
	var resp []HealthService
	if err := json.Unmarshal(res, &resp); err != nil {
		return nil, err
	}
	var services []Service
	for _, s := range resp {
		if svr := s.To(); svr != nil {
			services = append(services, *svr)
		}
	}
	return services, nil
}
