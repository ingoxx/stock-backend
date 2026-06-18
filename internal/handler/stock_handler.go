package handler

import (
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
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	inflowData, err := sh.svc.GetCapitalInflowData(false)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
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

	//queryParams := r.URL.Query()
	//code := queryParams.Get("code")
	//price := queryParams.Get("price")   // 委托价格
	//hold := queryParams.Get("quantity") // 买入数量
	//if code == "" || price == "" || hold == "" {
	//	utils.ResponseJSON(w, StockResponse{
	//		Code: 1001,
	//		Msg:  "required parameter 'code' or 'price', 'hold' is missing or empty.",
	//		Data: "",
	//	})
	//	return
	//}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.AverageDownData](body, sh.vd, func(r *domain.AverageDownData) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.GetStockRealTimeData(req)
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
	data, err := sh.svc.GetStockRealTimeListV2()
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

	sad, err := sh.svc.GetAverageDownList()
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1003,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	var md = make(map[string]interface{})
	md["hd"] = hd
	md["data"] = data
	md["ad"] = sad

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: md,
	})
}

func (sh *StockHandler) DelSelfSelectedStockHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[GeneralStockReq](body, sh.vd, func(r *GeneralStockReq) {
		r.Code = strings.TrimSpace(r.Code)
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := sh.svc.DelSelfSelectedStock(req.Code); err != nil {
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[UpdateStockDealStatusReq](body, sh.vd, func(r *UpdateStockDealStatusReq) {
		r.Code = strings.TrimSpace(r.Code)
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.UpdateStockDealStatus(req.Code, req.Status)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[AddHistoryTradeReq](body, sh.vd, func(r *AddHistoryTradeReq) {
		r.Code = strings.TrimSpace(r.Code)
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.AddHistoryTradeData(req.Code, req.TradeType)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[StockSwitchReq](body, sh.vd, func(r *StockSwitchReq) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.StockRealTimeInfoSwitch(req.Status)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1004,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	var msg string
	if req.Status == 1 {
		msg = "实时刷新已关闭"
	} else if req.Status == 2 {
		msg = "实时刷新已开启"
	} else {
		msg = "ok"
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  msg,
		Data: data,
	})
}

func (sh *StockHandler) GetStockRtDataHandler(w http.ResponseWriter, r *http.Request) {
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[StockSwitchReq](body, sh.vd, func(r *StockSwitchReq) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.StockNoticeSwitch(req.Status)
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
	queryParams := r.URL.Query()
	industry := queryParams.Get("industry")
	days := queryParams.Get("days")
	lookBackDays := queryParams.Get("lookBackDays")
	price := queryParams.Get("price")
	trend := queryParams.Get("trend")

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

	if trend == "" {
		trend = "2"
	}

	data, err := sh.svc.GetGoodStocks(industry, days, lookBackDays, price, trend)
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
	queryParams := r.URL.Query()
	trend := queryParams.Get("trend")
	if trend == "" {
		trend = "2"
	}

	data, err := sh.svc.FilterGoodStocksHistory(trend)
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

	req, err := bindAndValidate[FsNoticeConfigReq](body, sh.vd, func(r *FsNoticeConfigReq) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := sh.svc.StockNoticeFsSet(req.WebHook, req.Word); err != nil {
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

	req, err := bindAndValidate[UpdateStockHoldingsReq](body, sh.vd, func(r *UpdateStockHoldingsReq) {
		r.Code = strings.TrimSpace(r.Code)
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.UpdateStockHoldings(req.Code, req.Price, req.Quantity)
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

func (sh *StockHandler) GetSelfSelectedStockListHandler(w http.ResponseWriter, r *http.Request) {
	data, err := sh.svc.GetSelfSelectedStockList()
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

func (sh *StockHandler) AddSelfSelectedStockHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[GeneralStockReq](body, sh.vd, func(r *GeneralStockReq) {
		r.Code = strings.TrimSpace(r.Code)
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.AddSelfSelectedStock(req.Code)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
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

func (sh *StockHandler) SelfSelectedStockDelHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[GeneralStockReq](body, sh.vd, func(r *GeneralStockReq) {
		r.Code = strings.TrimSpace(r.Code)
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := sh.svc.SelfSelectedStockDel(req.Code); err != nil {
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

func (sh *StockHandler) SetAiApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.AiApiKey](body, sh.vd, func(r *domain.AiApiKey) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.SetAiApiKey(req)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
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

func (sh *StockHandler) GetAiApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	data, err := sh.svc.GetAiApiKey()
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

func (sh *StockHandler) SetAverageDownInfoHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.AverageDownData](body, sh.vd, func(r *domain.AverageDownData) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.SetAverageDownInfo(req)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
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

func (sh *StockHandler) SetStockTriggeringRulesAlertsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.TriggeringRules](body, sh.vd, func(r *domain.TriggeringRules) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := sh.svc.SetStockTriggeringRulesAlerts(req); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
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

func (sh *StockHandler) GetStockTriggeringRulesAlertsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := sh.svc.GetStockTriggeringRulesAlerts()
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

func (sh *StockHandler) SetStockTaggingHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.StockTaggingData](body, sh.vd, func(r *domain.StockTaggingData) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := sh.svc.SetStockTagging(req)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
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

func (sh *StockHandler) GetStockTaggingHandler(w http.ResponseWriter, r *http.Request) {
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

	data, err := sh.svc.GetStockTagging(code)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
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
