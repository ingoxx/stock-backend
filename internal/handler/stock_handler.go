package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/stock-backend/internal/domain"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/utils"
)

type StockHandler struct {
	svc *service.StockService
	vd  *validator.Validate
}

type StockResponse struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Other interface{} `json:"other,omitempty"`
}

type GeneralStockReq struct {
	Code string `json:"code" validate:"required"`
}

type UpdateStockDealStatusReq struct {
	Code   string `json:"code" validate:"required"`
	Status int    `json:"status" validate:"required"`
}

type AddHistoryTradeReq struct {
	Code      string `json:"code" validate:"required"`
	TradeType int    `json:"trade_type" validate:"required"`
}

type StockSwitchReq struct {
	Status int `json:"status" validate:"required"`
}

type GetGoodStockReq struct {
	Industry     string `json:"industry" validate:"required"`
	Days         int    `json:"days" validate:"required"`
	LookBackDays int    `json:"look_back_days" validate:"required"`
}

type FsNoticeConfigReq struct {
	WebHook string `json:"web_hook"`
	Word    string `json:"word"`
}

type UpdateStockHoldingsReq struct {
	Code     string  `json:"code" validate:"required"`
	Price    float64 `json:"price" validate:"required"`
	Quantity int     `json:"quantity" validate:"required"`
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
	ud, err := sh.svc.GetIndustryStockUpDown()
	if err != nil {
		http.Error(w, err.Error(), 1001)
		return
	}

	data, err := sh.svc.GetShIndexRealTimeData(false)
	if err != nil {
		http.Error(w, err.Error(), 1002)
		return
	}

	inflowData, err := sh.svc.GetCapitalInflowData(false)
	if err != nil {
		http.Error(w, err.Error(), 1002)
		return
	}

	md := make(map[string]interface{})
	md["data"] = ud
	md["feishu_set_status"] = sh.svc.CheckStockNoticeFsSetStatus()
	md["sh_index_data"] = data
	md["capital_inflow_data"] = inflowData

	utils.ResponseJSON(w, StockResponse{
		Code:  1000,
		Msg:   "ok",
		Data:  md,
		Other: sh.svc.CheckStockNoticeFsSetStatus(),
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
	days := queryParams.Get("days")
	if code == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'name' is missing or empty.",
			Data: "",
		})
		return
	}

	if days == "" {
		days = "30"
	}

	data, err := sh.svc.GetStockHistoryData(code, days)
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
	price := queryParams.Get("price")   // 委托价格
	hold := queryParams.Get("quantity") // 买入数量
	if code == "" || price == "" || hold == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'code' or 'price', 'hold' is missing or empty.",
			Data: "",
		})
		return
	}

	data, err := sh.svc.GetStockRealTimeData(code, price, hold)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	hd, err := sh.svc.GetHistoryTradeDataList()
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	var md = make(map[string][]*domain.StockInfo)
	md["hd"] = hd
	md["data"] = data

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: md,
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

	hd, err := sh.svc.GetHistoryTradeDataList()
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	var md = make(map[string][]*domain.StockInfo)
	md["hd"] = hd
	md["data"] = data

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: md,
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

	var ssd GeneralStockReq
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

func (sh *StockHandler) UpdateStockDealStatusHandler(w http.ResponseWriter, r *http.Request) {
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

	var ssd UpdateStockDealStatusReq
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

	data, err := sh.svc.UpdateStockDealStatus(ssd.Code, ssd.Status)
	if err != nil {
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
		Data: data,
	})
}

func (sh *StockHandler) AddHistoryTradeDataHandler(w http.ResponseWriter, r *http.Request) {
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

	var ssd AddHistoryTradeReq
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

	data, err := sh.svc.AddHistoryTradeData(ssd.Code, ssd.TradeType)
	if err != nil {
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
		Data: data,
	})
}

func (sh *StockHandler) GetHistoryTradeDataListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Invalid Method", 403)
		return
	}

	data, err := sh.svc.GetHistoryTradeDataList()
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

func (sh *StockHandler) StockRealTimeInfoSwitchHandler(w http.ResponseWriter, r *http.Request) {
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

	var ssd StockSwitchReq
	if err := json.Unmarshal(body, &ssd); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

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

	data, err := sh.svc.StockRealTimeInfoSwitch(ssd.Status)
	if err != nil {
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
		Data: data,
	})
}

func (sh *StockHandler) GetStockRtDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Invalid Method", 403)
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

	data, err := sh.svc.GetStockRtData(code)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})

}

func (sh *StockHandler) StockNoticeSwitchHandler(w http.ResponseWriter, r *http.Request) {
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

	var ssd StockSwitchReq
	if err := json.Unmarshal(body, &ssd); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

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

	data, err := sh.svc.StockNoticeSwitch(ssd.Status)
	if err != nil {
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
		Data: data,
	})
}

func (sh *StockHandler) GetGoodStocksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	queryParams := r.URL.Query()
	industry := queryParams.Get("industry")
	days := queryParams.Get("days")
	lookBackDays := queryParams.Get("lookBackDays")
	price := queryParams.Get("price")
	if industry == "" || days == "" || lookBackDays == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'industry', 'days', 'lookBackDays' is missing or empty.",
			Data: "",
		})
		return
	}

	if price == "" {
		price = "0.1"
	}

	data, err := sh.svc.GetGoodStocks(industry, days, lookBackDays, price)
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
		Msg:  fmt.Sprintf("正在分析%s行业下的所有股票数据,大概需要1分钟左右", industry),
		Data: data,
	})
}

func (sh *StockHandler) FilterGoodStocksHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", 403)
		return
	}

	data, err := sh.svc.FilterGoodStocksHistory()
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

func (sh *StockHandler) StockNoticeFsSetHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	var ssd FsNoticeConfigReq
	if err := json.Unmarshal(body, &ssd); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

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
			Code: 1004,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	if err := sh.svc.StockNoticeFsSet(ssd.WebHook, ssd.Word); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1005,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "配置完成",
		Data: "",
	})
}

func (sh *StockHandler) SendFsInfoHandler(w http.ResponseWriter, r *http.Request) {
	if err := sh.svc.SendFsInfo(); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1004,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "发送完成,请检查是否收到信息",
		Data: "",
	})
}

func (sh *StockHandler) GetStockHistoryDataDateRangeHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	start := queryParams.Get("start")
	end := queryParams.Get("end")
	if code == "" || start == "" || end == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'code' or 'start_date' or 'end_date' is missing or empty.",
			Data: "",
		})
		return
	}

	data, err := sh.svc.GetStockHistoryDataDateRange(code, start, end)
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

func (sh *StockHandler) UpdateStockHoldingsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	var ssd UpdateStockHoldingsReq
	if err := json.Unmarshal(body, &ssd); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

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

	data, err := sh.svc.UpdateStockHoldings(ssd.Code, ssd.Price, ssd.Quantity)
	if err != nil {
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
		Data: data,
	})
}

func (sh *StockHandler) GetShIndexRealTimeDataHandler(w http.ResponseWriter, r *http.Request) {
	data, err := sh.svc.GetShIndexRealTimeData(true)

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

func (sh *StockHandler) GetCapitalInflowDataHandler(w http.ResponseWriter, r *http.Request) {
	data, err := sh.svc.GetCapitalInflowData(true)
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
