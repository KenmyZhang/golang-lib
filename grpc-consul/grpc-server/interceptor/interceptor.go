package interceptor

import (
	"fmt"
	"github.com/KenmyZhang/golang-lib/middleware"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

func Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			middleware.GrpcErrCounter.WithLabelValues(info.FullMethod).Inc()

			flags := map[string]string{"type": "gRPC", "endpoint": info.FullMethod}
			fmt.Println("flags:", flags, ", err:", err)
		}
	}()

	startTime := time.Now()
	reply, err := handler(ctx, req)
	latency := time.Since(startTime)

	elapsed := float64(latency.Nanoseconds()) / 1000000
	middleware.GrpcLatency.WithLabelValues(info.FullMethod).Observe(elapsed)
	middleware.GrpcCounter.WithLabelValues(info.FullMethod).Inc()
	log.Info(fmt.Sprintf("[GRPC] %s | %13v | %#v %v", info.FullMethod, latency, req, err))

	return reply, err
}
