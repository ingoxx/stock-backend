package domain

type GoldenPrice string

type GoldenInfo struct {
	Date  string      `json:"date"`
	Price GoldenPrice `json:"price"`
}

type GoldenInfoRepository interface {
	GetGoldenPriceList() ([]*GoldenInfo, error)
	SetGoldenDiffPrice(price float64) error
	SetGoldenSellPrice(price float64) error
	SetGoldenBuyPrice(price float64) error
}
