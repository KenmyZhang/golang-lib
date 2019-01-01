package main

import (
	"context"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"os"
	"time"

	"fmt"
	pb "github.com/KenmyZhang/golang-lib/grpc-service/proto"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50053"
	defaultName = "world"

	SERVICE_NAME              = "simple_zipkin_server"
	ZIPKIN_HTTP_ENDPOINT      = "http://127.0.0.1:9411/api/v1/spans"
	ZIPKIN_RECORDER_HOST_PORT = "127.0.0.1:9000"
)

func main() {
	collector, err := zipkin.NewHTTPCollector(ZIPKIN_HTTP_ENDPOINT)
	if err != nil {
		log.Error(fmt.Sprintf("zipkin.NewHTTPCollector err: %v", err))
		return
	}

	recorder := zipkin.NewRecorder(collector, true, ZIPKIN_RECORDER_HOST_PORT, SERVICE_NAME)

	tracer, err := zipkin.NewTracer(recorder, zipkin.ClientServerSameSpan(false))
	if err != nil {
		log.Error(fmt.Sprintf("zipkin.NewTracer err: %v", err))
		return
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(
		otgrpc.OpenTracingClientInterceptor(tracer, otgrpc.LogPayloads()),
	))
	if err != nil {
		log.Error(fmt.Sprintf("did not connect: %v", err))
		return
	}
	defer conn.Close()
	c := pb.NewFetchServiceClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Fetch(ctx, &pb.FetchRequest{Name: name, Ids: []int64{234, 456}})
	if err != nil {
		log.Error(fmt.Sprintf("could not fetch: %v", err))
		return
	}
	log.Info(fmt.Sprintf("Greeting: %+v", r.Results))
}

// All future RPC activity involving `conn` will be automatically traced.
