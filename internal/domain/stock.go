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
	AccountId     string      `json:"accountId,omitempty"` // A股，港股，美股
	Ticktime      FormatTime  `json:"ticktime"`
	Per           float64     `json:"per"`
	Pb            float64     `json:"pb"`
	Mktcap        float64     `json:"mktcap"`
	Nmc           float64     `json:"nmc"`
	Price         float64     `json:"price,omitempty"` // 买入价格
	Turnoverratio float64     `json:"turnoverratio"`
	Volume        int         `json:"volume"`
	Amount        int         `json:"amount"`
	Quantity      int         `json:"quantity"`           // 持仓数量
	IsDeal        int         `json:"is_deal"`            // 是否已成交， 1：委托中，2：已成交
	SellOff       int         `json:"sell_off,omitempty"` // 是否卖出，1：持有/买入，2：卖出

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
}

//type StockInfoForDateRepository interface {
//	GetStockInfoForDataList() ([]*StockHistoryDate, error)
//}
