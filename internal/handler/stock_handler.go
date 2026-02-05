package handler

import (
	"encoding/json"
	"github.com/ingoxx/stock-backend/internal/service"
	"log"
	"net/http"
)

type StockHandler struct {
	svc *service.StockService
}

type StockResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewStockHandler(svc *service.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

func (sh *StockHandler) GetStockListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	list, err := sh.svc.GetStockList()
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	var resp = StockResponse{
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
