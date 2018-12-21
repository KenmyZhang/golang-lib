package redis_client

import (
	"github.com/KenmyZhang/golang-lib/metric"
	"github.com/go-redis/redis"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
	"time"
	"encoding/gob"
	"bytes"
)

var redisSupplier *RedisSupplier

type RedisSupplier struct {
	Client *redis.Client
}

func NewRedisSupplier(options *redis.Options) *RedisSupplier {
	redisSupplier = &RedisSupplier{}
	redisSupplier.Client = redis.NewClient(options)

	if val, err := redisSupplier.Client.Ping().Result(); err != nil {
		log.Error("Unable to ping redis server: " + err.Error())
		return nil
	} else {
		log.Info("new redis supplier success, " + val)
	}

	return redisSupplier
}

func redisStatCall(latency time.Duration, method string, err error) {
	metric.CacheLatency.WithLabelValues(method).Observe(latency.Seconds() * 1000.0)
	metric.CacheCounter.WithLabelValues(method).Inc()
	if nil != err && redis.Nil != err {
		metric.CacheMissCounter.WithLabelValues(method).Inc()
	}
}

func (s *RedisSupplier) Setex(key string, value interface{}, expiry time.Duration) error {
	if bytes, err := GetBytes(value); err != nil {
		return err
	} else {
		sTime := time.Now()
		res := s.Client.Set(key, bytes, expiry)
		eTime := time.Now()
		redisStatCall(eTime.Sub(sTime), "Setex", res.Err())
		if res.Err() != nil {
			return res.Err()
		}
	}
	return nil
}


func (s *RedisSupplier) Get(key string, value interface{}) error {
	sTime := time.Now()
	res, err := s.Client.Get(key).Bytes()
	eTime := time.Now()
	redisStatCall(eTime.Sub(sTime), "Get", err)

	if err != nil {
		return err
	} else {
		if err := DecodeBytes(res, value); err != nil {
			return err
		}
		return nil
	}
}


func (s *RedisSupplier) Del(key string) error {
	sTime := time.Now()
	res := s.Client.Del(key)
	eTime := time.Now()
	redisStatCall(eTime.Sub(sTime), "Del", res.Err())
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeBytes(input []byte, thing interface{}) error {
	dec := gob.NewDecoder(bytes.NewReader(input))
	return dec.Decode(thing)
}
