package handler

import (
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/utils"
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

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: list,
	})
}

func (sh *StockHandler) GetStockInfoForDataListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "request method error", 403)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "miss param", 400)
		return
	}

	list, err := sh.svc.GetStockInfoForDataList(code)
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: list,
	})
}

func (sh *StockHandler) GetStockIndustryListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	list, err := sh.svc.GetStockIndustryList()
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: list,
	})

}

func (sh *StockHandler) GetIndustryStockUpDownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	ud, err := sh.svc.GetIndustryStockUpDown()
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: ud,
	})
}

func (sh *StockHandler) GetStockMarketDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	ud, err := sh.svc.GetStockMarketData()
	if err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: ud,
	})

}

func (sh *StockHandler) GetStockDataSwitchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	if err := sh.svc.GetStockDataSwitch(); err != nil {
		http.Error(w, err.Error(), 200)
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "retrieving the latest data, pls wait a minute.",
		Data: "",
	})

}

func (sh *StockHandler) GetStockDataStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	if err := sh.svc.GetStockDataStatus(); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: "",
	})
}

func (sh *StockHandler) GetIndustryDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	queryParams := r.URL.Query()
	name := queryParams.Get("name")
	if name == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'name' is missing or empty.",
			Data: "",
		})
		return
	}

	data, err := sh.svc.GetIndustryData(name)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})
}

func (sh *StockHandler) GetStockCusDaysDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	if code == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'name' is missing or empty.",
			Data: "",
		})
		return
	}

	data, err := sh.svc.GetStockHistoryData(code)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})
}
