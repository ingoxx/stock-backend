package redis

import (
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/domain"
	"sync"
)

type StockRepo struct {
	mu     sync.RWMutex
	client *redis.Client
}

func NewStockRepo(client *redis.Client) domain.StockInfoRepository {
	return &StockRepo{
		client: client,
	}
}

func (sr *StockRepo) GetStockList() ([]*domain.StockInfo, error) {
	return nil, nil
}
