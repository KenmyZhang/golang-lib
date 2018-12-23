package main

import (
	"context"
	"fmt"
	"github.com/KenmyZhang/golang-lib/grpc-consul/consul"
	pb "github.com/KenmyZhang/golang-lib/grpc-consul/proto"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"time"
)

var (
	client pb.FetchServiceClient
)

const (
	consulAddr  = ":8500"
	serviceName = "my_grpc_service"
)

func main() {
	consulClient := consul.NewConsulClient(consulAddr)
	Connect(consulClient)

	for {
		time.Sleep(1 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		name := "hello world"
		r, err := client.Fetch(ctx, &pb.FetchRequest{Name: name, Ids: []int64{234, 456}})
		if err != nil {
			log.Error("could not fetch: " + err.Error())
		}
		log.Info(fmt.Sprintf("Greeting: %+v", r.Results))
	}

}

func Connect(c *consul.ConsulClient) {
	r, _ := manual.GenerateAndRegisterManualResolver()
	conn, err := grpc.Dial(r.Scheme()+":///grpc_service.server", grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		log.Error("Init gRPC client fail, error: " + err.Error())
	}
	client = pb.NewFetchServiceClient(conn)

	if c.Client != nil {
		services, _, err := c.Client.Catalog().Service(serviceName, "", nil)
		if err != nil {
			log.Error("Discover service  fail, error: " + err.Error())
		} else {
			if len(services) == 0 {
				log.Error("Discover service  not found")
			}
			addrs := make([]resolver.Address, 0, len(services))
			for _, s := range services {
				addrs = append(addrs, resolver.Address{Addr: fmt.Sprintf("%s:%d", s.ServiceAddress, s.ServicePort)})
			}
			r.NewAddress(addrs)
		}

		go func(r *manual.Resolver) {
			// watch service update
			var lastIndex uint64
			for {
				services, metaInfo, err := c.Client.Catalog().Service(
					serviceName, "", &api.QueryOptions{WaitIndex: lastIndex})
				if err != nil {
					log.Error("Discover service  fail, error: " + err.Error())
				} else {
					lastIndex = metaInfo.LastIndex
					if len(services) == 0 {
						log.Error("service  not found")
					}
					addrs := make([]resolver.Address, 0, len(services))
					for _, s := range services {
						addrs = append(addrs, resolver.Address{Addr: fmt.Sprintf("%s:%d", s.ServiceAddress, s.ServicePort)})
					}
					r.NewAddress(addrs)
				}
			}
		}(r)
	}
}
