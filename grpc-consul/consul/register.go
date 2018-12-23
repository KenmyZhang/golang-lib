package consul

import (
	"fmt"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"github.com/hashicorp/consul/api"
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

func (c *ConsulClient) Register(serviceName, ip string, port int) {
	if ip != "" {
		id := fmt.Sprintf("grpc-%s-%d", ip, port)
		log.Info("Register gRPC service " + id)
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
		log.Info("Deregister gRPC service " + id)
		err := c.Client.Agent().ServiceDeregister(id)
		if err != nil {
			log.Error(fmt.Sprintf("Deregister gRPC service fail, error: %v.", err))
		} else {
			log.Info("deregister grpc service success")
		}
	}
}
