package main

import (
	"context"
	"fmt"
	"github.com/KenmyZhang/golang-lib/grpc-consul/consul"
	"github.com/KenmyZhang/golang-lib/grpc-consul/grpc-server/interceptor"
	pb "github.com/KenmyZhang/golang-lib/grpc-consul/proto"
	"github.com/KenmyZhang/golang-lib/middleware"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

var (
	grpcServer *grpc.Server
	lis        net.Listener
)

const (
	port        = 50051
	serviceName = "my_grpc_service"
	consulAddr  = ":8500"
	ip          = "127.0.0.1"
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
	middleware.AppName = "grpc_service"
	log.Info("new server")
	logCfg := &log.LoggerConfiguration{
		EnableConsole: true,
		ConsoleJson:   true,
		ConsoleLevel:  "info",
		EnableFile:    true,
		FileJson:      true,
		FileLevel:     "info",
		FileLocation:  "./grpc.log",
		MaxSize:       30 * 1024,
		MaxAge:        3,
		MaxBackups:    3,
		LocalTime:     true,
		Compress:      true,
	}
	logger := log.NewLogger(logCfg)
	log.RedirectStdLog(logger)
	log.InitGlobalLogger(logger)

	router := gin.New()
	router.Use(middleware.Prometheus())
	router.Use(middleware.Recovery())
	router.GET(middleware.DefaultMetricPath, middleware.GetMetrics)
	router.GET("health", Health)

	wg := &sync.WaitGroup{}
	StartServer(wg, port)
	consulClient := consul.NewConsulClient(consulAddr)
	consulClient.Register(serviceName, ip, port)
	wg.Wait()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	<-ch
	consulClient.Deregister(ip, port)
	stopServer(wg)
	wg.Wait()

}

func StartServer(wg *sync.WaitGroup, port int) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info(fmt.Sprintf("Start grpc server :%d.", port))
		grpcServer = grpc.NewServer(grpc.UnaryInterceptor(interceptor.Interceptor))
		pb.RegisterFetchServiceServer(grpcServer, &server{})
		reflection.Register(grpcServer)

		var err error
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			panic(fmt.Sprintf("Listen to port %d fail, error: %v.", port, err))
		}

		err = grpcServer.Serve(lis)
		if err != nil {
			panic(fmt.Sprintf("Start grpc server fail, error: %v.", err))
		}

		log.Critical("Stop grpc server.")
	}()
}

func stopServer(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if grpcServer != nil {
			grpcServer.GracefulStop()
			lis.Close()
		}
	}()
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}
