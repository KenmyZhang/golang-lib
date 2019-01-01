package main

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"net"

	"fmt"
	pb "github.com/KenmyZhang/golang-lib/grpc-service/proto"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"strconv"
	"time"
)

const (
	port2 = ":50052"

	SERVICE_NAME2              = "simple_zipkin_server2"
	ZIPKIN_HTTP_ENDPOINT2      = "http://127.0.0.1:9411/api/v1/spans"
	ZIPKIN_RECORDER_HOST_PORT2 = "127.0.0.1:9000"
	address2                   = "localhost:50051"
	defaultName2               = "world"
)

type server2 struct{}

func (s *server2) Fetch(ctx context.Context, in *pb.FetchRequest) (*pb.FetchResponse, error) {
	//创建一个 Zipkin HTTP 后端收集器
	collector, err := zipkin.NewHTTPCollector(ZIPKIN_HTTP_ENDPOINT2)
	if err != nil {
		log.Error(fmt.Sprintf("zipkin.NewHTTPCollector err: %v", err))
		return nil, err
	}

	//创建一个基于 Zipkin 收集器的记录器
	recorder := zipkin.NewRecorder(collector, true, ZIPKIN_RECORDER_HOST_PORT2, SERVICE_NAME2)

	//创建一个 OpenTracing 跟踪器（兼容 Zipkin Tracer）
	tracer, err := zipkin.NewTracer(recorder, zipkin.ClientServerSameSpan(false))
	if err != nil {
		log.Error(fmt.Sprintf("zipkin.NewTracer err: %v", err))
		return nil, err
	}
	// Set up a connection to the server.
	conn, err := grpc.Dial(address2, grpc.WithInsecure(), grpc.WithUnaryInterceptor(
		otgrpc.OpenTracingClientInterceptor(tracer, otgrpc.LogPayloads()),
	))
	if err != nil {
		log.Error(fmt.Sprintf("did not connect: %v", err))
		return nil, err
	}
	defer conn.Close()
	c := pb.NewFetchServiceClient(conn)
	name := defaultName2
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Fetch(ctx, &pb.FetchRequest{Name: name, Ids: []int64{234, 456}})
	if err != nil {
		log.Error(fmt.Sprintf("could not fetch: %v", err))
		return nil, err
	}
	log.Info(fmt.Sprintf("Greeting: %+v", r.Results))

	log.Info(fmt.Sprintf("FetchRequest:%+v", in))
	ids := ""
	for _, val := range in.Ids {
		ids = ids + strconv.FormatInt(val, 10)
	}
	rst := &pb.FetchResponse{Results: []string{"name:" + in.Name, "ids:" + ids}}
	log.Info(fmt.Sprintf("FetchResponse:%+v", rst))
	return rst, nil
}

func main() {
	//创建一个 Zipkin HTTP 后端收集器
	collector, err := zipkin.NewHTTPCollector(ZIPKIN_HTTP_ENDPOINT2)
	if err != nil {
		log.Error(fmt.Sprintf("zipkin.NewHTTPCollector err: %v", err))
		return
	}

	//创建一个基于 Zipkin 收集器的记录器
	recorder := zipkin.NewRecorder(collector, true, ZIPKIN_RECORDER_HOST_PORT2, SERVICE_NAME2)

	//创建一个 OpenTracing 跟踪器（兼容 Zipkin Tracer）
	tracer, err := zipkin.NewTracer(recorder, zipkin.ClientServerSameSpan(false))
	if err != nil {
		log.Error(fmt.Sprintf("zipkin.NewTracer err: %v", err))
		return
	}

	opts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			//otgrpc.OpenTracingClientInterceptor 返回 grpc.UnaryServerInterceptor，
			// 不同点在于该拦截器会在 gRPC Metadata 中查找 OpenTracing SpanContext。如果找到则为该服务的 Span Context 的子节点
			otgrpc.OpenTracingServerInterceptor(tracer, otgrpc.LogPayloads()),
		),
	}

	lis, err := net.Listen("tcp", port2)
	if err != nil {
		log.Error(fmt.Sprintf("failed to listen: %v", err))
		return
	}
	s := grpc.NewServer(opts...)
	pb.RegisterFetchServiceServer(s, &server2{})
	// Register reflection service on gRPC server.
	log.Info(fmt.Sprintf("listen:%v", port2))
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Error(fmt.Sprintf("failed to serve: %v", err))
		return
	}
}

//总的来讲，就是初始化 Zipkin，其又包含收集器、记录器、跟踪器。再利用拦截器在 Server 端实现 SpanContext、Payload 的双向读取和管理
