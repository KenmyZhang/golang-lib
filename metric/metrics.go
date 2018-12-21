package metric

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var RespCounter *prometheus.CounterVec
var CacheCounter *prometheus.CounterVec
var ErrCounter *prometheus.CounterVec
var CacheMissCounter *prometheus.CounterVec
var RespLatency *prometheus.HistogramVec
var CacheLatency *prometheus.HistogramVec

func Init(appName string) {
	historyBuckets := [...]float64{
		10., 20., 30., 50., 80., 100., 200., 300., 500., 1000., 2000., 3000.}


	RespCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: appName,
		Name:      appName + "_requests_total",
		Help:      "Request counts"}, []string{"method", "endpoint"})

	ErrCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: appName,
		Name:      appName + "_error_total",
		Help:      "Error counts"}, []string{"method", "endpoint"})

	RespLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: appName,
		Name:      appName + "_resp_latency_millisecond",
		Help:      "Response latency (millisecond)",
		Buckets:   historyBuckets[:]}, []string{"method", "endpoint"})

	CacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: appName,
		Name:      appName + "_cache_total",
		Help:      "cache counts"}, []string{"method"})

	CacheMissCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: appName,
		Name:      appName + "_cache_miss_total",
		Help:      "Cache miss counts"}, []string{"method"})

	CacheLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: appName,
		Name:      appName + "_cache_latency_millisecond",
		Help:      "Cache latency (millisecond)",
		Buckets:   historyBuckets[:]}, []string{"method"})

	prometheus.MustRegister(RespCounter)
	prometheus.MustRegister(ErrCounter)
	prometheus.MustRegister(RespLatency)
	prometheus.MustRegister(CacheCounter)
	prometheus.MustRegister(CacheMissCounter)
	prometheus.MustRegister(CacheLatency)
}

func GetMetrics(c *gin.Context) {
	handler := promhttp.Handler()
	handler.ServeHTTP(c.Writer, c.Request)
}

