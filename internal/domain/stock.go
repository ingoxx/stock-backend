package domain

type FormatTime string

type StockInfo struct {
	Symbol        string     `json:"symbol"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Trade         string     `json:"trade"`
	Buy           string     `json:"buy"`
	Sell          string     `json:"sell"`
	Settlement    string     `json:"settlement"`
	Open          string     `json:"open"`
	High          string     `json:"high"`
	Low           string     `json:"low"`
	Ticktime      FormatTime `json:"ticktime"`
	PriceChange   float64    `json:"pricechange"`
	ChangePercent float64    `json:"changepercent"`
	Per           float64    `json:"per"`
	Pb            float64    `json:"pb"`
	Mktcap        float64    `json:"mktcap"`
	Nmc           float64    `json:"nmc"`
	Turnoverratio float64    `json:"turnoverratio"`
	Volume        int        `json:"volume"`
	Amount        int        `json:"amount"`
}

type StockInfoRepository interface {
	GetStockList() ([]*StockInfo, error)
}
