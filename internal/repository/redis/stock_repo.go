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
				return dss, err
			}

			dss = append(dss, &ds)
		}
	}

	return dss, nil
}

func (sr *StockRepo) GetStockInfoForDataList(code string) ([]*domain.StockInfoForDate, error) {
	var ds []*domain.StockInfoForDate
	key := "stock_every_day_detail"

	result, err := sr.client.HGet(key, code).Result()
	if err != nil {
		return ds, err
	}

	bn := bytes.NewBufferString(result)
	if err := json.Unmarshal(bn.Bytes(), &ds); err != nil {
		return ds, err
	}

	return ds, nil
}

func (sr *StockRepo) GetStockIndustryList() ([]*domain.StockIndustryMap, error) {
	var ds []*domain.StockIndustryMap
	key := "industry_map"

	result, err := sr.client.Get(key).Result()
	if err != nil {
		return ds, err
	}

	bn := bytes.NewBufferString(result)
	if err := json.Unmarshal(bn.Bytes(), &ds); err != nil {
		return ds, err
	}

	return ds, nil
}

func (sr *StockRepo) GetIndustryStockUpDown() ([]*domain.StockIndustryUpDown, error) {
	var ud []*domain.StockIndustryUpDown
	result, err := sr.client.Get("industry_stock_up_down").Result()
	if err != nil {
		return ud, err
	}

	bn := bytes.NewBufferString(result)
	if err := json.Unmarshal(bn.Bytes(), &ud); err != nil {
		return ud, err
	}

	return ud, nil

}

func (sr *StockRepo) GetStockMarketData() (domain.StockMarketData, error) {
	var md domain.StockMarketData
	result, err := sr.client.Get("market_data").Result()
	if err != nil {
		return md, err
	}

	bn := bytes.NewBufferString(result)
	if err := json.Unmarshal(bn.Bytes(), &md); err != nil {
		return md, err
	}

	return md, nil
}

func (sr *StockRepo) GetStockDataSwitch() error {
	if err := sr.client.Set("run_stock", 1, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (sr *StockRepo) GetStockDataStatus() error {
	result, err := sr.client.Get("run_stock").Result()
	if err != nil {
		return err
	}

	if result != "2" {
		return fmt.Errorf("still running")
	}

	return nil
}

func (sr *StockRepo) GetIndustryData(name string) ([]*domain.StockInfo, error) {
	var md []*domain.StockInfo

	result, err := sr.client.HGet("all_industry_data_ha", name).Result()
	if err != nil {
		return nil, err
	}

	bn := bytes.NewBufferString(result)
	if err := json.Unmarshal(bn.Bytes(), &md); err != nil {
		return nil, err
	}

	if md == nil {
		return nil, fmt.Errorf("fail to Unmarshal data")
	}

	fmt.Println(md)

	return md, nil
}
