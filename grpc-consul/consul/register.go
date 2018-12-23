package consul

import (
	"fmt"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"github.com/hashicorp/consul/api"
	"strings"
)

type ConsulClient struct {
	Client *api.Client
}

func NewConsulClient(consulAddr string) *ConsulClient {
	var err error
	cfg := &api.Config{Address: consulAddr}
	Client, err := api.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("Connect to consul fail, error: %v.", err))
	}
	return &ConsulClient{Client: Client}
}

func (c *ConsulClient) loadConfigFromConsul(prefix string) (cfg map[string]string) {
	cfg = make(map[string]string)
	pairs, _, err := c.Client.KV().List(prefix, nil)
	if err != nil {
		panic(fmt.Sprintf("Load config from Consul fail, error: %v", err))
	}
	for _, pair := range pairs {
		cfg[strings.Trim(strings.TrimPrefix(pair.Key, prefix), "/")] = string(pair.Value)
	}
	log.Info(fmt.Sprintf("Load config from Consul: %v.", cfg))
	return cfg
}

func (c *ConsulClient) Register(serviceName, ip string, port int) {
	if ip != "" {
		id := fmt.Sprintf("grpc-%s-%d", ip, port)
		log.Info("Register gRPC service %s." + id)
		reg := &api.AgentServiceRegistration{
			ID:      id,
			Name:    serviceName,
			Port:    port,
			Address: ip,
			Check: &api.AgentServiceCheck{
				HTTP:     fmt.Sprintf("http://%s:%d/health", ip, port),
				Interval: "1m",
			},
		}

		if err := c.Client.Agent().ServiceRegister(reg); err != nil {
			panic(fmt.Sprintf("Register gRPC service fail, error: %v.", err))
		}
	}
}

func (c *ConsulClient) Deregister(ip string, port int) {
	if ip != "" {
		id := fmt.Sprintf("grpc-%s-%d", ip, port)
		err := c.Client.Agent().ServiceDeregister(id)
		if err != nil {
			log.Error(fmt.Sprintf("Deregister gRPC service fail, error: %v.", err))
		}
	}
}
