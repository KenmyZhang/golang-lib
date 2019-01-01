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
)

const (
	port = ":50051"

	SERVICE_NAME              = "simple_zipkin_server"
	ZIPKIN_HTTP_ENDPOINT      = "http://127.0.0.1:9411/api/v1/spans"
	ZIPKIN_RECORDER_HOST_PORT = "127.0.0.1:9000"
)

type server struct{}

func (s *server) Fetch(ctx context.Context, in *pb.FetchRequest) (*pb.FetchResponse, error) {
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
	collector, err := zipkin.NewHTTPCollector(ZIPKIN_HTTP_ENDPOINT)
	if err != nil {
		log.Error(fmt.Sprintf("zipkin.NewHTTPCollector err: %v", err))
		return
	}

	//创建一个基于 Zipkin 收集器的记录器
	recorder := zipkin.NewRecorder(collector, true, ZIPKIN_RECORDER_HOST_PORT, SERVICE_NAME)

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

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Error(fmt.Sprintf("failed to listen: %v", err))
		return
	}
	s := grpc.NewServer(opts...)
	pb.RegisterFetchServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	log.Info(fmt.Sprintf("listen:%v", port))
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Error(fmt.Sprintf("failed to serve: %v", err))
		return
	}
}

//总的来讲，就是初始化 Zipkin，其又包含收集器、记录器、跟踪器。再利用拦截器在 Server 端实现 SpanContext、Payload 的双向读取和管理
