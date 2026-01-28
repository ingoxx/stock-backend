package handler

import (
	"encoding/json"
	"github.com/ingoxx/stock-backend/internal/service"
	"io"
	"log"
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
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (gh *GoldenHandler) SetGoldenPriceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", 403)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	var sg SetGoldenPriceRequest
	if err := json.Unmarshal(body, &sg); err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	if err := gh.svc.SetGoldenBuyPrice(sg.BuyPrice); err != nil {
		http.Error(w, err.Error(), 200)
		return
	}
	if err := gh.svc.SetGoldenDiffPrice(sg.DiffPrice); err != nil {
		http.Error(w, err.Error(), 200)
		return
	}
	if err := gh.svc.SetGoldenSellPrice(sg.SellPrice); err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	var resp = GoldenPriceResponse{
		Code: 1000,
		Msg:  "设置成功",
		Data: "",
	}

	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.Printf("%s, fail to response, '%s'", r.URL, err.Error())
	}
}

func (gh *GoldenHandler) GetGoldenPriceListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	list, err := gh.svc.GetGoldenPriceList()
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	var resp = GoldenPriceResponse{
		Code: 1000,
		Msg:  "ok",
		Data: list,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.Printf("%s, fail to response, '%s'", r.URL, err.Error())
	}

}
