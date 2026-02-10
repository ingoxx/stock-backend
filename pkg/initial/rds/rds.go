package rds

import (
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/configs"
	"sync"
	"time"
)

var (
	lock sync.RWMutex
)

func NewRedisClient(db int) (*redis.Client, error) {
	rds := redis.NewClient(
		&redis.Options{
			Addr:         configs.RedisHost,
			DB:           db,
			MinIdleConns: 5,
			Password:     configs.RedisPwd,
			PoolSize:     5,
			PoolTimeout:  30 * time.Second,
			DialTimeout:  30 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	)

	if err := rds.Ping(); err != nil {
		return rds, err.Err()
	}

	return rds, nil
}

func GetRedisClient(db int, rc map[int]*redis.Client) *redis.Client {
	lock.RLock()
	client, ok := rc[db]
	lock.RUnlock()

	if client != nil && ok {
		return client
	}

	lock.Lock()
	defer lock.Unlock()

	nClient, err := NewRedisClient(db)
	if err != nil {
		panic(err)
	}

	rc[db] = nClient

	return nClient
}
