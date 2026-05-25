package domain

import "encoding/json"

type FormatTime string

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
	Day    string  `json:"day"`
	Code   string  `json:"code"`
	Open   string  `json:"open"`
	High   string  `json:"high"`
	Low    string  `json:"low"`
	Volume string  `json:"volume"`
	PctChg float64 `json:"pct_chg"`
	Close  float64 `json:"close"`
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
	Code string   `json:"code"`
	Date []string `json:"date"`
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
	GetStockHistoryData(code string) ([]*StockHistoryDate, error)
	GetStockInfoData(code string) (*StockInfo, error)
	GetStockRealTimeData(code, price, hold string) ([]*StockInfo, error)
	GetStockRealTimeList() ([]*StockInfo, error)
	DelSelfSelectedStock(code string) error
	UpdateStockDealStatus(code string, status int) ([]*StockInfo, error)
	AddHistoryTradeData(code string, TradeType int) ([]*StockInfo, error)
	GetHistoryTradeDataList() ([]*StockInfo, error)
	StockRealTimeInfoSwitch(status int) (string, error)
	GetStockRtData(code string) (*StockInfo, error)
	StockNoticeSwitch(status int) (int, error)
	GetGoodStocks(industry, days, lookBackDays string) ([]*FilterGoodStock, error)
	// UpdateStockEntrustStatus(code string, status int) ([]*StockInfo, error)

}
