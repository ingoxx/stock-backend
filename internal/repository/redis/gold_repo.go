package redis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/domain"
	"sync"
)

type GoldRepo struct {
	mu     sync.RWMutex
	client *redis.Client
}

func NewGoldRepo(client *redis.Client) domain.GoldenInfoRepository {
	return &GoldRepo{
		client: client,
	}
}

func (gr *GoldRepo) GetGoldenPriceList() ([]*domain.GoldenInfo, error) {
	result, err := gr.client.LRange("real_time_golden_price", -10, -1).Result()
	if err != nil {
		return nil, err
	}

	var ds = make([]*domain.GoldenInfo, 0, len(result))
	for _, v := range result {
		var d domain.GoldenInfo

		b := bytes.NewBufferString(v)
		if err := json.Unmarshal(b.Bytes(), &d); err != nil {
			return ds, err
		}

		ds = append(ds, &d)
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
	gr.mu.Lock()
	defer gr.mu.Unlock()

	np := fmt.Sprintf("%f", price)

	return gr.client.Set("golden_pd", np, 0).Err()
}

func (gr *GoldRepo) SetGoldenSellPrice(price float64) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	np := fmt.Sprintf("%f", price)

	return gr.client.Set("sell_gold_price", np, 0).Err()
}

func (gr *GoldRepo) SetGoldenBuyPrice(price float64) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	np := fmt.Sprintf("%f", price)

	return gr.client.Set("buy_gold_price", np, 0).Err()
}
