package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

var RespCounter *prometheus.CounterVec
var ErrCounter *prometheus.CounterVec
var RespLatency *prometheus.HistogramVec

var CacheCounter *prometheus.CounterVec
var CacheMissCounter *prometheus.CounterVec
var CacheLatency *prometheus.HistogramVec

var GrpcCounter *prometheus.CounterVec
var GrpcErrCounter *prometheus.CounterVec
var GrpcLatency *prometheus.HistogramVec

const (
	DefaultMetricPath = "/metrics"
)

var AppName string

func init() {
	if AppName == "" {
		AppName = "app"
	}
	historyBuckets := [...]float64{
		10., 20., 30., 50., 80., 100., 200., 300., 500., 1000., 2000., 3000.}

	RespCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: AppName,
		Name:      AppName + "_requests_total",
		Help:      "Request counts"}, []string{"method", "endpoint"})

	ErrCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: AppName,
		Name:      AppName + "_error_total",
		Help:      "Error counts"}, []string{"method", "endpoint"})

	RespLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: AppName,
		Name:      AppName + "_resp_latency_millisecond",
		Help:      "Response latency (millisecond)",
		Buckets:   historyBuckets[:]}, []string{"method", "endpoint"})

	CacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: AppName,
		Name:      AppName + "_cache_total",
		Help:      "cache counts"}, []string{"method"})

	CacheMissCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: AppName,
		Name:      AppName + "_cache_miss_total",
		Help:      "Cache miss counts"}, []string{"method"})

	CacheLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: AppName,
		Name:      AppName + "_cache_latency_millisecond",
		Help:      "Cache latency (millisecond)",
		Buckets:   historyBuckets[:]}, []string{"method"})

	GrpcCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: AppName,
		Name:      AppName + "_grpc_total",
		Help:      "grpc call counts"}, []string{"method"})

	GrpcErrCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: AppName,
		Name:      AppName + "_grpc_error_total",
		Help:      "grpc error counts"}, []string{"method"})

	GrpcLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: AppName,
		Name:      AppName + "_grpc__latency_millisecond",
		Help:      "grpc latency (millisecond)",
		Buckets:   historyBuckets[:]}, []string{"method"})

	prometheus.MustRegister(RespCounter)
	prometheus.MustRegister(ErrCounter)
	prometheus.MustRegister(RespLatency)
	prometheus.MustRegister(CacheCounter)
	prometheus.MustRegister(CacheMissCounter)
	prometheus.MustRegister(CacheLatency)
	prometheus.MustRegister(GrpcCounter)
	prometheus.MustRegister(GrpcErrCounter)
	prometheus.MustRegister(GrpcLatency)
}

func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		endPoint := c.Request.URL.Path
		if endPoint == DefaultMetricPath {
			c.Next()
		} else {
			start := time.Now()
			method := c.Request.Method

			c.Next()
			relativePath := c.GetString("RELATIVE_PATH")
			if relativePath != "" {
				endPoint = relativePath
			}

			statusCode := c.Writer.Status()
			if statusCode != http.StatusNotFound {
				elapsed := float64(time.Since(start).Nanoseconds()) / 1000000
				RespCounter.WithLabelValues(method, endPoint).Inc()
				RespLatency.WithLabelValues(method, endPoint).Observe(elapsed)
			} else {
				elapsed := float64(time.Since(start).Nanoseconds()) / 1000000
				RespCounter.WithLabelValues(method, "[!othersPath!]").Inc()
				RespLatency.WithLabelValues(method, "[!othersPath!]").Observe(elapsed)
			}
		}
	}
}

func GetMetrics(c *gin.Context) {
	handler := promhttp.Handler()
	handler.ServeHTTP(c.Writer, c.Request)
}
