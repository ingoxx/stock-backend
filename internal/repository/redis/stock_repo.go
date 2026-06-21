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
	stockRealTimeSwitchKey       = "stock_real_time_switch"
	stockNoticeKey               = "stock_real_time_notice"
	filterGoodStockKey           = "filter_good_stock"
	stockFsSetKey                = "fei_bot"
	stockHistoryDataDateRangeKey = "stock_history_date_range"
	shIndexKey                   = "sh_index_rt"
	capitalInFlowKey             = "capital_inflow"
	stockRealTimeDataLockKey     = "lock:stock_real_time_data"
	aiSecretKey                  = "ai_api_key"
	selfSelectedStocksKey        = "self_selected_stocks"
	stockAverageDownKey          = "stock_average_down"
	stockBuyChangeKey            = "stock_buy_change"
	stockTriggeringRulesKey      = "stock_monitor_config"
	stockTaggingKey              = "stock_tagging"
	// 模拟的持仓数量最多20个，具体要看服务器配置
	maxMonitorStocks = 20
)

type StockRepo struct {
	mu     sync.RWMutex
	client *redis.Client
	wg     sync.WaitGroup
	// 恢复 singleflight，防止用户打开多个浏览器标签页导致的重复请求挤兑服务器
	sf singleflight.Group
}

func NewStockRepo(client *redis.Client) domain.StockInfoRepository {
	return &StockRepo{
		client: client,
	}
}

// withStockRealTimeDataLock 修改为【细粒度单支股票锁】
// 这就是让你原代码从“单线程卡顿”变成“双线程狂飙”的核心！不同股票更新再也不用互相排队等锁了！
func (sr *StockRepo) withStockRealTimeDataLock(code string, waitTimeout time.Duration, fn func() error) error {
	token, err := newLockToken()
	if err != nil {
		return err
	}

	const lockTTL = 2 * time.Minute

	// 按具体的股票代码加锁
	lockKey := stockRealTimeDataLockKey
	if code != "" {
		lockKey = fmt.Sprintf("%s:%s", stockRealTimeDataLockKey, code)
	}

	deadline := time.Now().Add(waitTimeout)
	for {
		ok, err := sr.client.SetNX(lockKey, token, lockTTL).Result()
		if err != nil {
			return err
		}
		if ok {
			break
		}

		// waitTimeout=0 表示非阻塞（后台刷新时碰到用户正在修改持仓，直接跳过保护性能）
		if waitTimeout == 0 {
			return fmt.Errorf("lock acquired by others, skip")
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
`, []string{lockKey}, token).Result()
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

	if err := json.Unmarshal([]byte(result), &md); err != nil {
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

func (sr *StockRepo) GetStockRealTimeData(data domain.AverageDownData) ([]*domain.StockInfo, error) {
	if err := sr.checkStockLimit(maxMonitorStocks); err != nil {
		return nil, err
	}

	return sr.getStockRealTimeDataV2(data)
}

func (sr *StockRepo) getStockRealTimeDataV2(data domain.AverageDownData) ([]*domain.StockInfo, error) {
	b, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}

	dl, err := sr.loadStockRealTimeData()
	if err != nil {
		return nil, err
	}

	var isExist bool
	var cd = new(domain.StockInfo)

	for _, v := range dl {
		if v.Code == data.Code {
			isExist = true
			cd = v
		}
	}

	if !isExist {
		if err := sr.client.HSet(stockBuyChangeKey, data.Code, string(b)).Err(); err != nil {
			return nil, err
		}
	}

	// 已经是持仓状态的才支持补仓
	if cd.IsDealStatus == 2 {
		if err := sr.client.HSet(stockAverageDownKey, data.Code, string(b)).Err(); err != nil {
			return nil, err
		}
		// 添加交易记录
		cd.Price = data.Price
		cd.Quantity = data.Quantity
		b2, err := json.Marshal(cd)
		if err != nil {
			return nil, err
		}

		if err := sr.client.RPush(stockTradeHistoryDataKey, string(b2)).Err(); err != nil {
			return nil, err
		}
	}

	return sr.loadStockRealTimeData()
}

func (sr *StockRepo) isInStockTime() bool {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	hour := now.Hour()
	minute := now.Minute()
	currentMinutes := hour*60 + minute

	isMorning := currentMinutes >= 570 && currentMinutes <= 700
	isAfternoon := currentMinutes >= 780 && currentMinutes <= 910

	return isMorning || isAfternoon
}

// GetStockRealTimeList 回归完全阻塞式调用，完美匹配前端等待机制
func (sr *StockRepo) GetStockRealTimeList() ([]*domain.StockInfo, error) {
	result, err := sr.client.Get(stockRealTimeSwitchKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
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

		// sf.Do 保证即使前端开10个网页，也只会真正触发一组更新任务，不浪费任何CPU性能
		_, err, _ = sr.sf.Do("refresh_stock_realtime_list", func() (interface{}, error) {
			var eg errgroup.Group
			// 严格控制只开2个并发，吃满双核但绝不造成堵塞崩溃
			eg.SetLimit(2)

			for _, item := range data {
				code := item.Code
				eg.Go(func() error {
					// 锁粒度变成基于 code，双核就可以真正同时跑2只不同的股票，效率翻倍！
					// waitTimeout=0 表示如果碰到用户正在修改，跳过不管。
					return sr.withStockRealTimeDataLock(code, 0, func() error {
						// 校验是否已被删除
						exists, err := sr.client.HExists(stockRealTimeDataKey, code).Result()
						if err == nil && !exists {
							return nil
						}
						// 同步调用，执行Python
						_ = sr.refreshStockRealTimeData(code, "", "")
						return nil
					})
				})
			}

			// 这里真正会阻塞，直到所有 Python 都执行完，完美适配前端的等待需求！
			_ = eg.Wait()
			return nil, nil
		})

		if err != nil {
			return nil, err
		}
	}

	// 等全部更新跑完后，重新加载 Redis 里最新的数据完整返回
	return sr.loadStockRealTimeData()
}

func (sr *StockRepo) GetStockRealTimeListV2() ([]*domain.StockInfo, error) {
	return sr.loadStockRealTimeData()
}

func (sr *StockRepo) checkStockLimit(limit int) error {
	current, err := sr.client.HGetAll(stockRealTimeDataKey).Result()
	if err != nil {
		return fmt.Errorf("get current stock data from redis: %w", err)
	}

	if len(current) >= limit {
		return fmt.Errorf("exceeding the position limit; maximum holding of %d stocks", limit)
	}

	return nil
}

func (sr *StockRepo) refreshStockRealTimeData(code, price, hold string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, pythonBin, pythonFile, code, price, hold)
	out, err := cmd.CombinedOutput()
	if err != nil {
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
		if err == nil && rd != nil {
			info.Industry = rd.Industry
			info.MainBusiness = rd.MainBusiness
		} else {
			info.Industry = "未知"
			info.MainBusiness = "未知"
		}
		result = append(result, &info)
	}

	return result, nil
}

func (sr *StockRepo) DelSelfSelectedStock(code string) error {
	return sr.withStockRealTimeDataLock(code, 30*time.Second, func() error {
		if err := sr.client.HDel(stockRealTimeDataKey, code).Err(); err != nil {
			return err
		}
		return nil
	})
}

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

func (sr *StockRepo) UpdateStockDealStatus(code string, status int) ([]*domain.StockInfo, error) {
	err := sr.withStockRealTimeDataLock(code, 30*time.Second, func() error {
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

	if len(result) == 0 {
		return nil, nil
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
	err := sr.withStockRealTimeDataLock(code, 30*time.Second, func() error {
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

func (sr *StockRepo) StockRealTimeInfoSwitch(status int) (string, error) {
	if err := sr.client.Set(stockRealTimeSwitchKey, status, 0).Err(); err != nil {
		return "", err
	}

	result, err := sr.client.Get(stockRealTimeSwitchKey).Result()
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

	key := fmt.Sprintf("%s_%s", filterGoodStockKey, trend)

	result, err := sr.client.HGet(key, industry).Result()
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

func (sr *StockRepo) FilterGoodStocksHistory(trend string) ([]string, error) {
	key := fmt.Sprintf("%s_%s", filterGoodStockKey, trend)
	result, err := sr.client.HKeys(key).Result()
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

func (sr *StockRepo) UpdateStockHoldings(code string, price float64, quantity int) ([]*domain.StockInfo, error) {
	err := sr.withStockRealTimeDataLock(code, 30*time.Second, func() error {
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

	// 1. 使用 HVals 直接获取所有的 JSON 字符串值
	result, err := sr.client.HVals(aiSecretKey).Result()
	if err != nil {
		// 注意：Redis 的 HVALS/HKEYS 命令在 Key 不存在时会返回空数组和 nil error，不会触发 redis.Nil。
		// 如果您想在没有配置时报错，可以在下方判断 len(result) == 0。
		return nil, fmt.Errorf("get ai api key: %w", err)
	}

	// 2. 如果没有任何配置参数，主动返回错误
	if len(result) == 0 {
		return nil, fmt.Errorf("missing AI API configuration parameters")
	}

	for _, v := range result {
		// 3. 实例化单条数据的指针
		sd := new(domain.AiApiKey)

		// 4. 将具体的 JSON 值 v 反序列化到单条对象 sd 中
		if err := json.Unmarshal([]byte(v), sd); err != nil {
			return nil, fmt.Errorf("unmarshal ai api key: %w", err)
		}

		// 5. 追加到结果切片中
		data = append(data, sd)
	}

	return data, nil
}

func (sr *StockRepo) SetAiApiKey(sk domain.AiApiKey) ([]*domain.AiApiKey, error) {
	b, err := json.Marshal(sk)
	if err != nil {
		return nil, fmt.Errorf("marshal ai api key: %w", err)
	}

	if err := sr.client.HSet(aiSecretKey, sk.Preset, string(b)).Err(); err != nil {
		return nil, fmt.Errorf("set ai api key: %s", err)
	}

	return sr.GetAiApiKey()
}

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

func (sr *StockRepo) SelfSelectedStockDel(code string) error {
	if err := sr.client.HDel(selfSelectedStocksKey, code).Err(); err != nil {
		return err
	}

	return nil
}

func (sr *StockRepo) GetAverageDownList() ([]*domain.AverageDownData, error) {
	var data []*domain.AverageDownData

	// 1. 使用 HVals 直接获取所有的 JSON 字符串值
	result, err := sr.client.HVals(stockAverageDownKey).Result()
	if err != nil {
		// 注意：Redis 的 HVALS/HKEYS 命令在 Key 不存在时会返回空数组和 nil error，不会触发 redis.Nil。
		// 如果您想在没有配置时报错，可以在下方判断 len(result) == 0。
		return nil, err
	}

	// 2. 如果没有任何配置参数，主动返回错误
	if len(result) == 0 {
		return data, nil
	}

	for _, v := range result {
		// 3. 实例化单条数据的指针
		sd := new(domain.AverageDownData)

		// 4. 将具体的 JSON 值 v 反序列化到单条对象 sd 中
		if err := json.Unmarshal([]byte(v), sd); err != nil {
			return nil, fmt.Errorf("unmarshal %s key: %w", stockAverageDownKey, err)
		}

		// 5. 追加到结果切片中
		data = append(data, sd)
	}

	return data, nil
}

func (sr *StockRepo) SetAverageDownInfo(ad domain.AverageDownData) ([]*domain.AverageDownData, error) {
	b, err := json.Marshal(ad)
	if err != nil {
		return nil, fmt.Errorf("marshal AverageDownData: %w", err)
	}

	if err := sr.client.HSet(stockAverageDownKey, ad.Code, string(b)).Err(); err != nil {
		return nil, err
	}

	return sr.GetAverageDownList()
}

func (sr *StockRepo) SetStockTriggeringRulesAlerts(rd domain.TriggeringRules) error {
	b, err := json.Marshal(&rd)
	if err != nil {
		return fmt.Errorf("marshal TriggeringRulesAlerts error: %w", err)
	}

	if err := sr.client.HSet(stockTriggeringRulesKey, rd.Code, string(b)).Err(); err != nil {
		return err
	}

	return nil
}

func (sr *StockRepo) GetStockTriggeringRulesAlerts() ([]*domain.TriggeringRules, error) {
	var data []*domain.TriggeringRules

	result, err := sr.client.HVals(stockTriggeringRulesKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get ai api key: %w", err)
	}

	if len(result) == 0 {
		return data, nil
	}

	for _, v := range result {
		sd := new(domain.TriggeringRules)

		if err := json.Unmarshal([]byte(v), sd); err != nil {
			return nil, fmt.Errorf("unmarshal ai api key: %w", err)
		}

		data = append(data, sd)
	}

	return data, nil
}

func (sr *StockRepo) SetStockTagging(data domain.StockTaggingData) (domain.StockTaggingData, error) {
	b, err := json.Marshal(&data)
	if err != nil {
		return data, err
	}

	if err := sr.client.HSet(stockTaggingKey, data.Code, string(b)).Err(); err != nil {
		return data, err
	}

	return data, nil
}

func (sr *StockRepo) GetStockTagging() ([]domain.StockTaggingData, error) {

	result, err := sr.client.HVals(stockTaggingKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	var data = make([]domain.StockTaggingData, 0, len(result))

	for _, v := range result {
		var sd domain.StockTaggingData
		if err := json.Unmarshal([]byte(v), &sd); err != nil {
			return nil, fmt.Errorf("unmarshal %s key: %w", stockTaggingKey, err)
		}

		data = append(data, sd)
	}

	return data, nil
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
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("run realtime script timeout: %s", string(out))
		}
		return fmt.Errorf("run realtime script failed: %w, output: %s", err, string(out))
	}

	return nil
}
