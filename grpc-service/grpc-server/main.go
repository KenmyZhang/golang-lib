package main

import (
	"context"
	"net"

	"google.golang.org/grpc"
	pb "github.com/KenmyZhang/golang-lib/grpc-service/proto"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"google.golang.org/grpc/reflection"
	"strconv"
	"fmt"
)

const (
	port = ":50051"
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
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Error(fmt.Sprintf("failed to listen: %v", err))
		return
	}
	s := grpc.NewServer()
	pb.RegisterFetchServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	log.Info(fmt.Sprintf("listen:%v", port))
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Error(fmt.Sprintf("failed to serve: %v", err))
		return
	}
}
