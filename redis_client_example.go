package main

import (
	"time"
	"github.com/go-redis/redis"
	"fmt"
	redisClient "github.com/KenmyZhang/golang-lib/redis-client"
	"github.com/KenmyZhang/golang-lib/metric"
	log "github.com/KenmyZhang/golang-lib/zaplogger"
)

func main() {

	metric.Init("example")

	options := 	&redis.Options{
		Addr:         "127.0.0.1:6379",
		Password:     "",
		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     30,
		PoolTimeout:  3 * time.Second,
		DB:           0,
	}

	redisSupplier := redisClient.NewRedisSupplier(options)

	var val int64

	if err := redisSupplier.Get("test1", &val);  err == nil {
		log.Info(fmt.Sprintf("redisSupplier.Get success, rst:%+v", val))
	} else if err == redis.Nil {
		log.Error(fmt.Sprintf("redisSupplier.Get success, val is nil"))
	} else {
		log.Error(fmt.Sprintf("redisSupplier.Get error, err:%+v", err))
	}

	val = 123456

	if err := redisSupplier.Setex("test1", &val, -1); err != nil {
		log.Error(fmt.Sprintf("redisSupplier.Setex, err:%+v", err))
	} else {
		log.Info(fmt.Sprintf("redisSupplier.Setex success, :%+v", val))
	}

	var rst int64

	if err := redisSupplier.Get("test1", &rst);  err == nil {
		log.Info(fmt.Sprintf("redisSupplier.Get success, rst:%+v", rst))
	} else {
		log.Error(fmt.Sprintf("redisSupplier.Get error, err:%+v", err))
	}

	if err := redisSupplier.Del("test1"); err == nil {
		log.Info(fmt.Sprintf("redisSupplier.Del success"))
	} else {
		log.Error(fmt.Sprintf("redisSupplier.Del error, err:%+v", err))
	}


	var rst2 int64

	if err := redisSupplier.Get("test1", &rst2);  err == nil {
		log.Info(fmt.Sprintf("redisSupplier.Get success, rst:%+v", rst2))
	} else if err == redis.Nil {
		log.Error(fmt.Sprintf("redisSupplier.Get success, val is nil"))
	} else {
		log.Error(fmt.Sprintf("redisSupplier.Get error, err:%+v", err))
	}
}
