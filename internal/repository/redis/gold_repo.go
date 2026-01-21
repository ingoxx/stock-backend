package redis

import (
	"github.com/ingoxx/stock-backend/internal/domain"
	"github.com/redis/go-redis"
	"sync"
)

type GoldRepo struct {
	mu     sync.RWMutex
	client *redis.Client
}

func NewGoldRepo(client *redis.Client) domain.GoldenRepository {
	return &GoldRepo{
		client: client,
	}
}

func (gr *GoldRepo) GetGoldenInfo() ([]*domain.GoldenInfo, error) {
	return nil, nil
}

func (gr *GoldRepo) SetGoldenDiffPrice(price float64) error {
	return nil
}

func (gr *GoldRepo) SetGoldenSellPrice(price float64) error {
	return nil
}

func (gr *GoldRepo) SetGoldenBuyPrice(price float64) error {
	return nil
}
