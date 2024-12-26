package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alackfeng/datacenter-bridge/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdRegistry - etcd服务发现客户端.
type EtcdRegistry struct {
	client     *clientv3.Client
	leaseID    clientv3.LeaseID
	prefix     string
	serviceTTL int64
	regKeys    sync.Map // // zone_service -> map[prefix_zone_service][]Service.
	self       Service  // 自己注册的服务.
}

var _ Discovery = (*EtcdRegistry)(nil)

// NewEtcdRegistry -
func NewEtcdRegistry(endpoints []string, prefix string, serviceTTL int64) *EtcdRegistry {
	if client, err := clientv3.New(clientv3.Config{Endpoints: endpoints,
		DialTimeout: 5 * time.Second}); err != nil {
		return nil
	} else {
		return &EtcdRegistry{
			client:     client,
			prefix:     prefix,
			serviceTTL: serviceTTL,
			regKeys:    sync.Map{},
		}
	}
}

func (e *EtcdRegistry) ID() string {
	return e.self.String()
}

func (e *EtcdRegistry) registerKey(service Service) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", e.prefix, service.Service, service.Zone, service.Id, service.Host)
}

func (e *EtcdRegistry) serviceKey(serviceName, zone string) string {
	return fmt.Sprintf("%s/%s/%s", e.prefix, serviceName, zone)
}

// func (e *EtcdRegistry) registerGet(key string) []Service {
// 	if v, ok := e.regKeys.Load(key); ok {
// 		return v.([]Service)
// 	}
// 	return nil
// }

func (e *EtcdRegistry) registerAdd(service Service) {
	key := e.serviceKey(service.Service, service.Zone)
	if v, ok := e.regKeys.Load(key); ok {
		services := v.([]Service)
		services = append(services, service)
		e.regKeys.Store(key, services)
	} else {
		e.regKeys.Store(key, []Service{service})
	}
}

func (e *EtcdRegistry) registerDel(key string) {
	var ks string
	var ns []Service
	e.regKeys.Range(func(k, v interface{}) bool {
		services := v.([]Service)
		for i, s := range services {
			if e.registerKey(s) == key {
				ks = k.(string)
				ns = append(services[:i], services[i+1:]...)
				return false
			}
		}
		return true
	})
	if ks != "" {
		e.regKeys.Store(ks, ns)
	}
}

// Register - implements Discovery.
func (e *EtcdRegistry) Register(ctx context.Context, service Service) error {
	if leaseResp, err := e.client.Grant(ctx, e.serviceTTL); err != nil {
		return err
	} else {
		e.leaseID = leaseResp.ID
		key := e.registerKey(service)
		value, err := json.Marshal(service)
		if err != nil {
			return err
		}
		// put key with lease.
		if _, err := e.client.Put(ctx, key, string(value), clientv3.WithLease(e.leaseID)); err != nil {
			return err
		}
		// keep lease alive.
		keepAliveCh, err := e.client.KeepAlive(ctx, e.leaseID)
		if err != nil {
			return err
		}
		e.self = service       // update self.
		e.registerAdd(service) // update reg keys.

		go func() {
			for {
				select {
				case <-keepAliveCh:
					fmt.Println("keep alive.")
				case <-ctx.Done():
					return
				}
			}
		}()

	}
	return nil
}

// Unregister -
func (e *EtcdRegistry) Unregister(ctx context.Context) error {
	if e.self.Id == "" {
		return nil
	}
	logger.Debugf("etcd unregister: %v", e.self)
	key := e.registerKey(e.self)
	if _, err := e.client.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

// Watch -
func (e *EtcdRegistry) Watch(ctx context.Context) {
	watchChan := e.client.Watch(ctx, e.prefix, clientv3.WithPrefix())
	go func() {
		for resp := range watchChan {
			for _, ev := range resp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					// key := string(ev.Kv.Key)
					var service Service
					if err := json.Unmarshal(ev.Kv.Value, &service); err != nil {
						logger.Warnf("etcd %v service put umarshal: %s, err %v", e.self, string(ev.Kv.Key), err)
						continue
					}
					logger.Infof("etcd %v service added/updated: %v, %v", e.self, string(ev.Kv.Key), service)
					e.registerAdd(service) // update reg keys.
				case clientv3.EventTypeDelete:
					logger.Infof("etcd %v service removed: %s\n", e.self, string(ev.Kv.Key))
					e.registerDel(string(ev.Kv.Key))
				}
			}
		}
	}()
}

// GetServices - implements Discovery.
func (e *EtcdRegistry) GetServices(ctx context.Context, zone string, serviceName string) ([]Service, error) {
	// cache get ?.
	key := e.serviceKey(serviceName, zone)
	// if s := e.registerGet(key); s != nil {
	// 	return s, nil
	// }
	resp, err := e.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	services := make([]Service, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var service Service
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}
