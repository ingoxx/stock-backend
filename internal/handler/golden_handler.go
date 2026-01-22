package handler

import (
	"encoding/json"
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

type GoldenPriceListResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (gh *GoldenHandler) SetGoldenPriceHandler(w http.ResponseWriter, r *http.Request) {

}

func (gh *GoldenHandler) GetGoldenPriceListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	if sign := r.FormValue("sign"); sign != "lqmlxb" {
		http.Error(w, "", 403)
		return
	}

	list, err := gh.svc.GetGoldenPriceList()
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	var resp = GoldenPriceListResponse{
		Code: 1000,
		Msg:  "ok",
		Data: list,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	w.Write(b)

}
