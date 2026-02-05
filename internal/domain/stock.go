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

type IndustryStockInfo struct {
	Name string       `json:"name"`
	Data []*StockInfo `json:"data"`
}

type StockInfoRepository interface {
	GetStockList() ([]*StockInfo, error)
}
