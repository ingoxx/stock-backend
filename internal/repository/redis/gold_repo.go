package redis

import (
	"encoding/json"
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

func (gr *GoldRepo) GetGoldenPriceList() ([]*domain.GoldenInfo, error) {
	result, err := gr.client.LRange("real_time_golden_price", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var ds = make([]*domain.GoldenInfo, 0, len(result))

	for _, v := range result {
		var d *domain.GoldenInfo
		if err := json.Unmarshal([]byte(v), d); err != nil {
			return ds, err
		}

		ds = append(ds, d)
	}

	return ds, nil
}

func (gr *GoldRepo) GetGoldenPriceRealTime() (string, error) {
	result, err := gr.client.Get("golden-real-time-price").Result()
	if err != nil {
		return result, err
	}

	return result, nil
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
