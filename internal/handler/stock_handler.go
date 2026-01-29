package handler

import (
	"github.com/ingoxx/stock-backend/internal/service"
	"net/http"
)

type StockHandler struct {
	svc *service.StockService
}

func NewStockHandler(svc *service.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

func (sh *StockHandler) GetStockListHandler(w http.ResponseWriter, r *http.Request) {}
