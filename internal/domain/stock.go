package domain

import (
	"encoding/json"
	"math"
	"strconv"
)

type FormatTime string

type FormatVolume string

func (fv FormatVolume) FormatFloat(volumeStr string) (float64, error) {
	num, err := strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return 0, err
	}

	return math.Round((num/100000000)*100) / 100, nil
}

type StockInfo struct {
	PriceChange   json.Number `json:"pricechange"`
	ChangePercent json.Number `json:"changepercent"`
	Symbol        string      `json:"symbol"`
	Code          string      `json:"code"`
	Name          string      `json:"name"`
	Trade         string      `json:"trade"`
	Buy           string      `json:"buy"`
	Sell          string      `json:"sell"`
	Settlement    string      `json:"settlement"`
	Open          string      `json:"open"`
	High          string      `json:"high"`
	Low           string      `json:"low"`
	Industry      string      `json:"industry"`
	MainBusiness  string      `json:"main_business"`
	AccountId     string      `json:"accountId"` // A股，港股，美股
	Ticktime      FormatTime  `json:"ticktime"`
	Per           float64     `json:"per"`
	Pb            float64     `json:"pb"`
	Mktcap        float64     `json:"mktcap"`
	Nmc           float64     `json:"nmc"`
	Price         float64     `json:"price"` // 买入价格
	Turnoverratio float64     `json:"turnoverratio"`
	ProfitLoss    float64     `json:"profit_loss"` // 具体的盈亏价格
	Bep           float64     `json:"bep"`         // 盈亏百分比
	Volume        int         `json:"volume"`
	Amount        int         `json:"amount"`
	Quantity      int         `json:"quantity"`        // 持仓数量
	IsDealStatus  int         `json:"is_deal_status"`  // 状态显示是否已成交， 1：委托中，2：已成交
	SellOffStatus int         `json:"sell_off_status"` // 状态显示是否卖出，1：持有/买入，2：卖出
	TradeType     int         `json:"trade_type"`      // 交易类型，1：已卖出，2：已撤回，3：买入
}

type StockHistoryDate struct {
	Day    string       `json:"day"`
	Code   string       `json:"code"`
	Open   string       `json:"open"`
	High   string       `json:"high"`
	Low    string       `json:"low"`
	Volume FormatVolume `json:"volume"`
	PctChg float64      `json:"pct_chg"`
	Close  float64      `json:"close"`
}

type StockIndustryMap struct {
	Name  string `json:"name"`
	Match int    `json:"match"`
}

type StockIndustryUpDown struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Up     int32  `json:"up"`
	Down   int32  `json:"down"`
}

type StockMarketData struct {
	Total  int   `json:"total"`
	Amount int   `json:"amount"`
	Up     int32 `json:"up"`
	Down   int32 `json:"down"`
}

type FilterGoodStock struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Date string `json:"date"`
}

// FsNoticeConfig 飞书预警配置
type FsNoticeConfig struct {
	WebHook string `json:"web_hook"`
	Word    string `json:"word"`
}

// CapitalInflow 实时的资金流向
type CapitalInflow struct {
	IndustryCode         string  `json:"industry_code"`
	IndustryName         string  `json:"industry_name"`
	IndustryIndex        float64 `json:"industry_index"`
	ChangePercent        float64 `json:"change_percent"`
	NetInflowBillionYuan float64 `json:"net_inflow_billion_yuan"`
	NetInflowPercent     float64 `json:"net_inflow_percent"`
}

// ShIndex 上证指数的数据结构
type ShIndex struct {
	Name          string  `json:"name"`
	Date          string  `json:"date"`
	Time          string  `json:"time"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	PrevClose     float64 `json:"prev_close"`
	ChangeAmount  float64 `json:"change_amount"`
	ChangePercent float64 `json:"change_percent"`
	Amount        float64 `json:"amount"`
	Volume        int     `json:"volume"`
}

type AiApiKey struct {
	Mode   string `json:"mode"`
	Preset string `json:"preset"`
	ApiKey string `json:"apiKey"`
	ApiUrl string `json:"apiUrl"`
	Model  string `json:"model"`
}

type AverageDownData struct {
	Code     string  `json:"code"  validate:"required"`
	Price    float64 `json:"price"  validate:"required"`
	Quantity int     `json:"quantity"  validate:"required"`
}

type StockInfoRepository interface {
	GetStockList() ([]*StockInfo, error)
	GetStockInfoForDataList(code string) ([]*StockHistoryDate, error)
	GetStockIndustryList() ([]*StockIndustryMap, error)
	GetIndustryStockUpDown() ([]*StockIndustryUpDown, error)
	GetStockMarketData() (StockMarketData, error)
	GetStockDataSwitch() error
	GetStockDataStatus() error
	GetIndustryData(name string) ([]*StockInfo, error)
	GetStockHistoryData(code, days string) ([]*StockHistoryDate, error)
	GetStockInfoData(code string) (*StockInfo, error)
	GetStockRealTimeData(data AverageDownData) ([]*StockInfo, error)
	GetStockRealTimeList() ([]*StockInfo, error)
	GetStockRealTimeListV2() ([]*StockInfo, error)
	DelSelfSelectedStock(code string) error
	UpdateStockDealStatus(code string, status int) ([]*StockInfo, error)
	AddHistoryTradeData(code string, TradeType int) ([]*StockInfo, error)
	GetHistoryTradeDataList() ([]*StockInfo, error)
	StockRealTimeInfoSwitch(status int) (string, error)
	GetStockRtData(code string) (*StockInfo, error)
	StockNoticeSwitch(status int) (int, error)
	GetGoodStocks(industry, days, lookBackDays, price, trend string) ([]*FilterGoodStock, error)
	FilterGoodStocksHistory(trend string) ([]string, error)
	StockNoticeFsSet(webHook, word string) error
	CheckStockNoticeFsSetStatus() bool
	SendFsInfo() error
	GetStockHistoryDataDateRange(code, start, end string) ([]*StockHistoryDate, error)
	UpdateStockHoldings(code string, price float64, quantity int) ([]*StockInfo, error)
	GetCapitalInflowData(isRun bool) ([]*CapitalInflow, error)
	GetShIndexRealTimeData(isRun bool) (*ShIndex, error)
	AddSelfSelectedStock(code string) ([]*StockInfo, error)
	UpdateSelfSelectedStock(code string) ([]*StockInfo, error)
	GetSelfSelectedStockList() ([]*StockInfo, error)
	SelfSelectedStockDel(code string) error
	SetAiApiKey(sk AiApiKey) ([]*AiApiKey, error)
	GetAiApiKey() ([]*AiApiKey, error)
	GetAverageDownList() ([]*AverageDownData, error)
	SetAverageDownInfo(ad AverageDownData) ([]*AverageDownData, error)
}
