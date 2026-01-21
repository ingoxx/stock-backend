package rds

import (
	"github.com/ingoxx/stock-backend/configs"
	"github.com/redis/go-redis"
	"time"
)

var (
	Rds *redis.Client
)

func init() {
	Rds = redis.NewClient(
		&redis.Options{
			Addr:         configs.RedisHost,
			DB:           configs.RedisDb,
			MinIdleConns: 5,
			Password:     configs.RedisPwd,
			PoolSize:     5,
			PoolTimeout:  30 * time.Second,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	)

	if err := Rds.Ping(); err != nil {
		panic(err)
	}
}
