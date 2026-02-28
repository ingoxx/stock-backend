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
	Ticktime      FormatTime  `json:"ticktime"`
	Per           float64     `json:"per"`
	Pb            float64     `json:"pb"`
	Mktcap        float64     `json:"mktcap"`
	Nmc           float64     `json:"nmc"`
	Turnoverratio float64     `json:"turnoverratio"`
	Volume        int         `json:"volume"`
	Amount        int         `json:"amount"`
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
}

//type StockInfoForDateRepository interface {
//	GetStockInfoForDataList() ([]*StockHistoryDate, error)
//}
