package handler

import (
	"github.com/ingoxx/stock-backend/internal/service"
	"net/http"
)

type GoldenHandler struct {
	svc *service.GoldenService
}

func NewGoldenHandler(svc *service.GoldenService) *GoldenHandler {
	return &GoldenHandler{svc: svc}
}

type SetGoldenPriceRequest struct {
	DiffPrice float64 `json:"diff_price"`
	BuyPrice  float64 `json:"buy_price"`
	SellPrice float64 `json:"sell_price"`
}

type GoldenPriceResponse struct {
}

func (gh *GoldenHandler) SetGoldenPriceHandler(w http.ResponseWriter, r *http.Request) {

}

func (gh *GoldenHandler) GetGoldenPriceHandler(w http.ResponseWriter, r *http.Request) {}
