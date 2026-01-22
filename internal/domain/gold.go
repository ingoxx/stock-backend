package domain

import "fmt"

type GoldenPrice float64

func (g GoldenPrice) Display() string {
	return fmt.Sprintf("¥%.2f/g", g)
}

type GoldenInfo struct {
	Date  string
	Price GoldenPrice
}

type GoldenRepository interface {
	GetGoldenPriceList() ([]*GoldenInfo, error)
	SetGoldenDiffPrice(price float64) error
	SetGoldenSellPrice(price float64) error
	SetGoldenBuyPrice(price float64) error
	GetGoldenPriceRealTime() (string, error)
}
