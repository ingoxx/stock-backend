package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/domain"
	"golang.org/x/sync/errgroup"
)

const (
	pythonBin  = "/usr/local/python3.10/bin/python3.10"
	pythonFile = "/root/pyscript/spot/stock_data_real_time.py"
)

type StockRepo struct {
	mu     sync.RWMutex
	client *redis.Client
	wg     sync.WaitGroup
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

func (sr *StockRepo) GetStockInfoForDataList(code string) ([]*domain.StockHistoryDate, error) {
	var ds []*domain.StockHistoryDate
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

	return md, nil
}

func (sr *StockRepo) GetStockHistoryData(code string) ([]*domain.StockHistoryDate, error) {
	var md []*domain.StockHistoryDate

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "/usr/local/python3.10/bin/python3.10", "/root/pyscript/spot/stock_history_data.py", code, "30")
	if err := command.Run(); err != nil {
		return md, err
	}

	result, err := sr.client.HGet("stock_every_day_detail", code).Result()
	if err != nil {
		return md, err
	}

	bn := bytes.NewBufferString(result)
	if err := json.Unmarshal(bn.Bytes(), &md); err != nil {
		return nil, err
	}

	return md, nil
}

func (sr *StockRepo) GetStockInfoData(code string) (*domain.StockInfo, error) {
	var keys = []string{"sh_a", "sz_a"}
	var data string
	var ds *domain.StockInfo

	for _, k := range keys {
		result, err := sr.client.HGet(k, code).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}

			return nil, err
		}

		if result != "" {
			data = result
			break
		}
	}

	if data == "" {
		return nil, fmt.Errorf("%s not found", code)
	}

	if err := json.Unmarshal([]byte(data), &ds); err != nil {
		return nil, err
	}

	return ds, nil
}

// GetStockRealTimeData 实时获取某个行情数据
func (sr *StockRepo) GetStockRealTimeData(code string) ([]*domain.StockInfo, error) {
	const maxStocks = 10

	if err := sr.checkStockLimit(maxStocks); err != nil {
		return nil, err
	}

	if err := sr.refreshStockRealTimeData(code); err != nil {
		return nil, err
	}

	return sr.loadStockRealTimeData()
}

// GetStockRealTimeList 从列表中获取每个最新行情数据
func (sr *StockRepo) GetStockRealTimeList() ([]*domain.StockInfo, error) {
	data, err := sr.loadStockRealTimeData()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return data, nil
	}

	var eg errgroup.Group
	eg.SetLimit(10)

	for _, item := range data {
		code := item.Code

		eg.Go(func() error {
			if err := sr.refreshStockRealTimeData(code); err != nil {
				return fmt.Errorf("refresh %s failed: %w", code, err)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return sr.loadStockRealTimeData()
}

func (sr *StockRepo) checkStockLimit(limit int) error {
	const redisKey = "stock_real_time_data"

	current, err := sr.client.HGetAll(redisKey).Result()
	if err != nil {
		return fmt.Errorf("get current stock data from redis: %w", err)
	}
	if len(current) >= limit {
		return fmt.Errorf("up to %d self-selected stocks", limit)
	}

	return nil
}

func (sr *StockRepo) refreshStockRealTimeData(code string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, pythonBin, pythonFile, code)

	out, err := cmd.CombinedOutput() // stdout + stderr
	if err != nil {
		// 超时要单独判断，便于定位
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("run realtime script timeout: %s", string(out))
		}
		return fmt.Errorf("run realtime script failed: %w, output: %s", err, string(out))
	}
	return nil
}

func (sr *StockRepo) loadStockRealTimeData() ([]*domain.StockInfo, error) {
	const redisKey = "stock_real_time_data"

	rawMap, err := sr.client.HGetAll(redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get latest stock data from redis: %w", err)
	}

	result := make([]*domain.StockInfo, 0, len(rawMap))
	for _, raw := range rawMap {
		var info domain.StockInfo
		if err := json.Unmarshal([]byte(raw), &info); err != nil {
			return nil, fmt.Errorf("unmarshal stock info: %w", err)
		}

		rd, err := sr.GetStockInfoData(info.Code)
		if err != nil {
			return nil, err
		}

		info.Industry = rd.Industry
		info.MainBusiness = rd.MainBusiness

		result = append(result, &info)
	}
	return result, nil
}

func (sr *StockRepo) DelSelfSelectedStock(code string) error {
	const redisKey = "stock_real_time_data"
	if err := sr.client.HDel(redisKey, code).Err(); err != nil {
		return err
	}

	return nil
}
