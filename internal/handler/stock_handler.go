package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/utils"
)

type StockHandler struct {
	svc *service.StockService
	vd  *validator.Validate
}

type StockResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type DelSelfSelectedStockReq struct {
	Code string `json:"code" validate:"required"`
}

func NewStockHandler(svc *service.StockService, vd *validator.Validate) *StockHandler {
	return &StockHandler{svc: svc, vd: vd}
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

func (sh *StockHandler) GetStockInfoDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	if code == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'code' is missing or empty.",
			Data: "",
		})
		return
	}

	data, err := sh.svc.GetStockInfoData(code)
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

func (sh *StockHandler) GetStockRealTimeDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	if code == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'code' is missing or empty.",
			Data: "",
		})
		return
	}

	data, err := sh.svc.GetStockRealTimeData(code)
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

func (sh *StockHandler) GetStockRealTimeListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	data, err := sh.svc.GetStockRealTimeList()
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

func (sh *StockHandler) DelSelfSelectedStockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", 403)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	var ssd DelSelfSelectedStockReq
	if err := json.Unmarshal(body, &ssd); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	ssd.Code = strings.TrimSpace(ssd.Code)

	if err := sh.vd.Struct(ssd); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, e := range validationErrors {
				utils.ResponseJSON(w, StockResponse{
					Code: 1003,
					Msg:  fmt.Sprintf("required parameter '%s' is missing or empty.", e.Field()),
					Data: "",
				})
				return
			}
		}

		utils.ResponseJSON(w, StockResponse{
			Code: 1003,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	if err := sh.svc.DelSelfSelectedStock(ssd.Code); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1004,
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
