package redis

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/domain"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

const (
	pythonBin            = "/usr/local/python3.10/bin/python3.10"
	pythonFile           = "/root/pyscript/spot/stock_data_real_time.py"
	stockRtDataFile      = "/root/pyscript/spot/get_stock_real_time.py"
	filterGoodStockFile  = "/root/pyscript/spot/filter_good_stock.py"
	stockHistoryDataFile = "/root/pyscript/spot/stock_history_data_day_by_day.py"
	feishuSendTestFile   = "/root/pyscript/spot/stock_notice.py"
	shIndexFile          = "/root/pyscript/spot/get_stock_sh_index_rt_data.py"
	capitalInFlowFile    = "/root/pyscript/spot/get_stock_capital_inflow_data.py"
	// redis key
	stockRealTimeDataKey         = "stock_real_time_data"
	stockTradeHistoryDataKey     = "stock_trade_history_data"
	stockRtDataKey               = "get_stock_rt_data"
	stockRealTimeSwitch          = "stock_real_time_switch"
	stockNoticeKey               = "stock_real_time_notice"
	filterGoodStockKey           = "filter_good_stock"
	stockFsSetKey                = "fei_bot"
	stockHistoryDataDateRangeKey = "stock_history_date_range"
	shIndexKey                   = "sh_index_rt"
	capitalInFlowKey             = "capital_inflow"
	stockRealTimeDataLockKey     = "lock:stock_real_time_data"
	aiSecretKey                  = "ai_api_key"
	selfSelectedStocksKey        = "self_selected_stocks"
)

type StockRepo struct {
	mu     sync.RWMutex
	client *redis.Client
	wg     sync.WaitGroup
	sf     singleflight.Group
}

func NewStockRepo(client *redis.Client) domain.StockInfoRepository {
	return &StockRepo{
		client: client,
	}
}

func (sr *StockRepo) withStockRealTimeDataLock(fn func() error) error {
	token, err := newLockToken()
	if err != nil {
		return err
	}

	const (
		lockTTL  = 2 * time.Minute
		lockWait = 30 * time.Second
	)

	deadline := time.Now().Add(lockWait)
	for {
		ok, err := sr.client.SetNX(stockRealTimeDataLockKey, token, lockTTL).Result()
		if err != nil {
			return err
		}
		if ok {
			break
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("stock realtime data is busy")
		}

		time.Sleep(100 * time.Millisecond)
	}

	defer func() {
		_, _ = sr.client.Eval(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`, []string{stockRealTimeDataLockKey}, token).Result()
	}()

	return fn()
}

func newLockToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
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

	if err := json.Unmarshal([]byte(result), &ds); err != nil {
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

func (sr *StockRepo) GetStockHistoryData(code, days string) ([]*domain.StockHistoryDate, error) {
	var md []*domain.StockHistoryDate

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "/usr/local/python3.10/bin/python3.10", "/root/pyscript/spot/stock_history_data.py", code, days)
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
func (sr *StockRepo) GetStockRealTimeData(code, price, hold string) ([]*domain.StockInfo, error) {
	const maxMonitorStocks = 20

	var data []*domain.StockInfo
	err := sr.withStockRealTimeDataLock(func() error {
		if err := sr.checkStockLimit(maxMonitorStocks); err != nil {
			return err
		}

		if err := sr.refreshStockRealTimeData(code, price, hold); err != nil {
			return err
		}

		var err error
		data, err = sr.loadStockRealTimeData()
		return err
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (sr *StockRepo) isInStockTime() bool {
	now := time.Now()

	// 获取当前星期 (time.Monday �?1, time.Sunday �?0)
	// �?Go 中，周一到周五对�?weekday 1 �?5
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// 获取当前时间的小时和分钟
	hour := now.Hour()
	minute := now.Minute()

	// 将时间转换为分钟数，方便比较 (例如 9:30 = 9*60 + 30 = 570)
	currentMinutes := hour*60 + minute

	// 上午 9:30 - 11:40 -> 570 - 700
	// 下午 13:00 - 15:10 -> 780 - 910

	isMorning := currentMinutes >= 570 && currentMinutes <= 700
	isAfternoon := currentMinutes >= 780 && currentMinutes <= 910

	return isMorning || isAfternoon
}

// GetStockRealTimeList returns latest realtime data
func (sr *StockRepo) GetStockRealTimeList() ([]*domain.StockInfo, error) {
	result, err := sr.client.Get(stockRealTimeSwitch).Result()
	if err != nil {
		return sr.loadStockRealTimeData()
	}

	if sr.isInStockTime() || result == "2" {

		data, err := sr.loadStockRealTimeData()
		if err != nil {
			return nil, err
		}
		if len(data) == 0 {
			return data, nil
		}

		_, err, _ = sr.sf.Do("refresh_stock_realtime_list", func() (interface{}, error) {
			var eg errgroup.Group
			eg.SetLimit(2)

			for _, item := range data {
				code := item.Code

				eg.Go(func() error {
					return sr.withStockRealTimeDataLock(func() error {
						if err := sr.refreshStockRealTimeData(code, "", ""); err != nil {
							return fmt.Errorf("refresh %s failed: %w", code, err)
						}
						return nil
					})
				})
			}

			if err := eg.Wait(); err != nil {
				return nil, err
			}
			return nil, nil
		})

		if err != nil {
			return nil, err
		}
	}

	return sr.loadStockRealTimeData()
}

func (sr *StockRepo) checkStockLimit(limit int) error {
	current, err := sr.client.HGetAll(stockRealTimeDataKey).Result()
	if err != nil {
		return fmt.Errorf("get current stock data from redis: %w", err)
	}

	if len(current) >= limit {
		return fmt.Errorf("up to %d self-selected stocks", limit)
	}

	return nil
}

func (sr *StockRepo) refreshStockRealTimeData(code, price, hold string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, pythonBin, pythonFile, code, price, hold)
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
	rawMap, err := sr.client.HGetAll(stockRealTimeDataKey).Result()
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
	return sr.withStockRealTimeDataLock(func() error {
		if err := sr.client.HDel(stockRealTimeDataKey, code).Err(); err != nil {
			return err
		}

		return nil
	})
}

// StockNoticeSwitch 1,close;2,open;3,check;default 1
func (sr *StockRepo) StockNoticeSwitch(status int) (int, error) {
	if status == 1 || status == 2 {
		if err := sr.client.Set(stockNoticeKey, status, 0).Err(); err != nil {
			return 0, err
		}

		return status, nil
	} else if status == 3 {
		result, err := sr.client.Get(stockNoticeKey).Result()
		if errors.Is(err, redis.Nil) {
			if err := sr.client.Set(stockNoticeKey, 1, 0).Err(); err != nil {
				return 1, err
			}
		}

		if err != nil {
			return 0, err
		}

		val, err := strconv.Atoi(result)
		if err != nil {
			return 0, err
		}

		return val, nil
	}

	return status, nil
}

// UpdateStockDealStatus updates trading status
func (sr *StockRepo) UpdateStockDealStatus(code string, status int) ([]*domain.StockInfo, error) {
	err := sr.withStockRealTimeDataLock(func() error {
		result, err := sr.client.HGet(stockRealTimeDataKey, code).Result()
		if errors.Is(err, redis.Nil) || result == "" {
			return nil
		}

		var info domain.StockInfo
		if err := json.Unmarshal([]byte(result), &info); err != nil {
			return fmt.Errorf("unmarshal stock info: %w", err)
		}

		now := time.Now()
		info.TradeType = status
		info.Ticktime = domain.FormatTime(now.Format("2006-01-02 15:04:05"))

		b, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("marshal stock info: %w", err)
		}

		if err := sr.client.HSet(stockRealTimeDataKey, code, string(b)).Err(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sr.GetStockRealTimeList()
}

func (sr *StockRepo) GetHistoryTradeDataList() ([]*domain.StockInfo, error) {
	result, err := sr.client.LRange(stockTradeHistoryDataKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var data = make([]*domain.StockInfo, 0, len(result))
	for _, raw := range result {
		var info domain.StockInfo
		if err := json.Unmarshal([]byte(raw), &info); err != nil {
			return nil, fmt.Errorf("fail to unmarshal stock info: %w", err)
		}
		data = append(data, &info)
	}

	return data, nil
}

func (sr *StockRepo) AddHistoryTradeData(code string, TradeType int) ([]*domain.StockInfo, error) {
	err := sr.withStockRealTimeDataLock(func() error {
		result, err := sr.client.HGet(stockRealTimeDataKey, code).Result()
		if errors.Is(err, redis.Nil) || result == "" {
			return nil
		}

		var info domain.StockInfo
		if err := json.Unmarshal([]byte(result), &info); err != nil {
			return fmt.Errorf("unmarshal stock info: %w", err)
		}

		info.TradeType = TradeType

		b, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("fail to marshal stock info: %w", err)
		}

		if err := sr.client.RPush(stockTradeHistoryDataKey, string(b)).Err(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sr.GetHistoryTradeDataList()
}

// StockRealTimeInfoSwitch 1=close, 2=open, 3=check, default 1
func (sr *StockRepo) StockRealTimeInfoSwitch(status int) (string, error) {
	if err := sr.client.Set(stockRealTimeSwitch, status, 0).Err(); err != nil {
		return "", err
	}

	result, err := sr.client.Get(stockRealTimeSwitch).Result()
	if err != nil {
		return result, err
	}

	return result, nil
}

func (sr *StockRepo) GetStockRtData(code string) (*domain.StockInfo, error) {
	if err := sr.runScript(stockRtDataFile, false, code); err != nil {
		return nil, err
	}

	var data *domain.StockInfo

	result, err := sr.client.HGet(stockRtDataKey, code).Result()
	if err != nil {
		return nil, err
	}

	if result == "" {
		return nil, fmt.Errorf(" %s stock info not found", code)
	}

	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("unmarshal stock info: %w", err)
	}

	return data, nil
}

func (sr *StockRepo) GetGoodStocks(industry, days, lookBackDays, price, trend string) ([]*domain.FilterGoodStock, error) {
	var data []*domain.FilterGoodStock

	if days != "1000" {
		if err := sr.runScript(filterGoodStockFile, true, industry, days, lookBackDays, price, trend); err != nil {
			return nil, err
		}
		return data, nil
	}

	result, err := sr.client.HGet(filterGoodStockKey, industry).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return data, nil
		}

		return nil, err
	}

	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("unmarshal stock info: %w", err)
	}

	return data, nil
}

func (sr *StockRepo) FilterGoodStocksHistory() ([]string, error) {
	result, err := sr.client.HKeys(filterGoodStockKey).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (sr *StockRepo) StockNoticeFsSet(webHook, word string) error {
	var s = domain.FsNoticeConfig{
		WebHook: webHook,
		Word:    word,
	}

	jsonStr, err := json.Marshal(&s)
	if err != nil {
		return err
	}

	if err := sr.client.Set(stockFsSetKey, string(jsonStr), 0).Err(); err != nil {
		return err
	}

	return nil
}

func (sr *StockRepo) SendFsInfo() error {
	if !sr.CheckStockNoticeFsSetStatus() {
		return errors.New("please set up Lark Robot first")
	}

	if err := sr.runScript(feishuSendTestFile, false); err != nil {
		return err
	}

	return nil
}

func (sr *StockRepo) CheckStockNoticeFsSetStatus() bool {
	result, err := sr.client.Get(stockFsSetKey).Result()
	if err != nil || result == "" || errors.Is(err, redis.Nil) {
		return false
	}

	return true
}

func (sr *StockRepo) GetStockHistoryDataDateRange(code, start, end string) ([]*domain.StockHistoryDate, error) {
	var md []*domain.StockHistoryDate

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, pythonBin, stockHistoryDataFile, code, start, end)
	if err := command.Run(); err != nil {
		return md, err
	}

	result, err := sr.client.HGet(stockHistoryDataDateRangeKey, code).Result()
	if err != nil {
		return md, err
	}

	if err := json.Unmarshal([]byte(result), &md); err != nil {
		return nil, err
	}

	return md, nil
}

// UpdateStockHoldings 修改持仓列表中的股票
func (sr *StockRepo) UpdateStockHoldings(code string, price float64, quantity int) ([]*domain.StockInfo, error) {
	err := sr.withStockRealTimeDataLock(func() error {
		result, err := sr.client.HGet(stockRealTimeDataKey, code).Result()
		if errors.Is(err, redis.Nil) || result == "" {
			return nil
		}

		var info domain.StockInfo
		if err := json.Unmarshal([]byte(result), &info); err != nil {
			return fmt.Errorf("unmarshal stock info: %w", err)
		}

		now := time.Now()
		info.Price = price
		info.Quantity = quantity
		info.Ticktime = domain.FormatTime(now.Format("2006-01-02 15:04:05"))

		b, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("marshal stock info: %w", err)
		}

		if err := sr.client.HSet(stockRealTimeDataKey, code, string(b)).Err(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sr.GetStockRealTimeList()
}

func (sr *StockRepo) GetShIndexRealTimeData(isRun bool) (*domain.ShIndex, error) {
	if isRun {
		if err := sr.runScript(shIndexFile, false); err != nil {
			return nil, err
		}
	}

	result, err := sr.client.HGet(shIndexKey, "sh").Result()
	if err != nil {
		return nil, err
	}

	var data *domain.ShIndex
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (sr *StockRepo) GetCapitalInflowData(isRun bool) ([]*domain.CapitalInflow, error) {
	if isRun {
		if err := sr.runScript(capitalInFlowFile, false); err != nil {
			return nil, err
		}
	}

	result, err := sr.client.HGet(capitalInFlowKey, "cf").Result()
	if err != nil {
		return nil, err
	}

	var data []*domain.CapitalInflow
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (sr *StockRepo) GetAiApiKey() ([]*domain.AiApiKey, error) {
	var data []*domain.AiApiKey
	result, err := sr.client.HKeys(aiSecretKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("missing AI API configuration parameters")
		}

		return nil, fmt.Errorf("get ai api key: %w", err)
	}

	for _, v := range result {
		var sd *domain.AiApiKey
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return nil, fmt.Errorf("unmarshal ai api key: %w", err)
		}
		data = append(data, sd)
	}

	return data, nil
}

// SetAiApiKey 设置ai的api key
func (sr *StockRepo) SetAiApiKey(sk map[string]interface{}) ([]*domain.AiApiKey, error) {
	var data []*domain.AiApiKey
	var sd *domain.AiApiKey

	marshal, err := json.Marshal(&sk)
	if err != nil {
		return nil, fmt.Errorf("marshal ai api key: %w", err)
	}

	if err := json.Unmarshal(marshal, sd); err != nil {
		return nil, fmt.Errorf("unmarshal ai api key: %w", err)
	}

	allData, err := sr.GetAiApiKey()
	if err != nil {
		return nil, fmt.Errorf("get ai api key: %w", err)
	}

	allData = append(allData, sd)
	jsonStr, err := json.Marshal(&sd)
	if err != nil {
		return nil, fmt.Errorf("marshal ai api key: %w", err)
	}

	if err := sr.client.HSet(aiSecretKey, sd.Name, string(jsonStr)); err != nil {
		return nil, fmt.Errorf("set ai api key: %s", err.Err())
	}

	return data, nil
}

// findStockByCode 自选stock的筛选
func (sr *StockRepo) findStockByCode(code string) (*domain.StockInfo, error) {
	stocks, err := sr.GetStockList()
	if err != nil {
		return nil, err
	}

	for _, item := range stocks {
		if item != nil && item.Code == code {
			return item, nil
		}
	}

	return nil, fmt.Errorf("%s not found", code)
}

// saveSelfSelectedStock 自选stock的update,add
func (sr *StockRepo) saveSelfSelectedStock(code string) ([]*domain.StockInfo, error) {
	stock, err := sr.findStockByCode(code)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(stock)
	if err != nil {
		return nil, fmt.Errorf("marshal self selected stock failed: %w", err)
	}

	if err := sr.client.HSet(selfSelectedStocksKey, code, string(b)).Err(); err != nil {
		return nil, err
	}

	return sr.GetSelfSelectedStockList()
}

func (sr *StockRepo) AddSelfSelectedStock(code string) ([]*domain.StockInfo, error) {
	return sr.saveSelfSelectedStock(code)
}

func (sr *StockRepo) UpdateSelfSelectedStock(code string) ([]*domain.StockInfo, error) {
	return sr.saveSelfSelectedStock(code)
}

// GetSelfSelectedStock 暂时用不到
func (sr *StockRepo) GetSelfSelectedStock(code string) (*domain.StockInfo, error) {
	result, err := sr.client.HGet(selfSelectedStocksKey, code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%s not found", code)
		}
		return nil, err
	}

	if result == "" {
		return nil, fmt.Errorf("%s not found", code)
	}

	var data domain.StockInfo
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (sr *StockRepo) GetSelfSelectedStockList() ([]*domain.StockInfo, error) {
	result, err := sr.client.HGetAll(selfSelectedStocksKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	data := make([]*domain.StockInfo, 0, len(result))
	for _, raw := range result {
		var info domain.StockInfo
		if err := json.Unmarshal([]byte(raw), &info); err != nil {
			return nil, err
		}
		data = append(data, &info)
	}

	return data, nil
}

// SelfSelectedStockDel 自选stock的del
func (sr *StockRepo) SelfSelectedStockDel(code string) error {
	if err := sr.client.HDel(selfSelectedStocksKey, code).Err(); err != nil {
		return err
	}

	return nil
}

func (sr *StockRepo) runScript(fileName string, async bool, args ...interface{}) error {
	cmdArgs := make([]string, 0, len(args)+2)
	cmdArgs = append(cmdArgs, pythonBin, fileName)
	for _, arg := range args {
		cmdArgs = append(cmdArgs, fmt.Sprint(arg))
	}

	if async {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start realtime script failed: %w", err)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
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
