package redis

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	var keys = []string{"sh_a", "sz_a"}

	var dss = make([]*domain.StockInfo, 0, 5200)

	for _, v := range keys {
		result, err := sr.client.HGetAll(v).Result()
		if err != nil {
			return dss, err
		}

		for m := range result {
			var ds domain.StockInfo
			s := result[m]
			bn := bytes.NewBufferString(s)
			if err := json.Unmarshal(bn.Bytes(), &ds); err != nil {
				fmt.Println("ERR >>> ", err, result[m])
				return dss, err
			}
			dss = append(dss, &ds)
		}
	}

	return dss, nil
}
