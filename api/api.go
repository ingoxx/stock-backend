package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/didip/tollbooth"
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/cmd/server"
	"github.com/ingoxx/stock-backend/config"
	"github.com/ingoxx/stock-backend/internal/middleware"
)

func Start() {
	var rdbConn = make(map[int]*redis.Client)

	goldenApp := server.NewGoldenApp(rdbConn)
	stockApp := server.NewStockApp(rdbConn)
	verifyApp := server.NewVerifyApp(rdbConn)
	docApp := server.NewDocApp()

	lmt := tollbooth.NewLimiter(config.MaxReqFrequency, nil)

	// 需要走中间件验证
	privateMux := http.NewServeMux()
	// gold变动接口
	privateMux.HandleFunc("GET /v2/golden/list", tollbooth.LimitFuncHandler(lmt, goldenApp.GoldenHandler.GetGoldenPriceListHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/golden/set", tollbooth.LimitFuncHandler(lmt, goldenApp.GoldenHandler.SetGoldenPriceHandler).ServeHTTP)

	// stock接口
	privateMux.HandleFunc("GET /v1/stock/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockListHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/days/detail", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockInfoForDataListHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/industry/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockIndustryListHandler).ServeHTTP)
	// 一并返回飞书预警信息是否已经配置
	privateMux.HandleFunc("GET /v1/stock/industry/up-down", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetIndustryStockUpDownHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/market-data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockMarketDataHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/switch", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockDataSwitchHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/run-status", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockDataStatusHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/industry/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetIndustryDataHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/history/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockCusDaysDataHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/info/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockInfoDataHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/real-time/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockRealTimeDataHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/real-time/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockRealTimeListHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/self-selected-del", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.DelSelfSelectedStockHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/trade-status/update", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.UpdateStockDealStatusHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/trade-history/add", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.AddHistoryTradeDataHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/trade-history/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetHistoryTradeDataListHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/real-time/switch", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.StockRealTimeInfoSwitchHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v2/stock/real-time/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockRtDataHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/notice/switch", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.StockNoticeSwitchHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/filter/good", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetGoodStocksHandler).ServeHTTP)
	// 筛选回调会上涨所有结果
	privateMux.HandleFunc("GET /v1/stock/filter/good/history", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.FilterGoodStocksHistoryHandler).ServeHTTP)
	// 自选stock list
	privateMux.HandleFunc("GET /v1/stock/self-selected/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetSelfSelectedStockListHandler).ServeHTTP)
	// add 自选stock
	privateMux.HandleFunc("POST /v1/stock/self-selected/add", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.AddSelfSelectedStockHandler).ServeHTTP)
	// del 自选stock
	privateMux.HandleFunc("POST /v2/stock/self-selected/del", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SelfSelectedStockDelHandler).ServeHTTP)

	// 获取上证指数
	privateMux.HandleFunc("GET /v1/stock/sh-index", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetShIndexRealTimeDataHandler).ServeHTTP)
	// 获取资金流向
	privateMux.HandleFunc("GET /v1/stock/capital-inflow", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetCapitalInflowDataHandler).ServeHTTP)
	// 修改持仓列表中的股票成本以及股数
	privateMux.HandleFunc("POST /v1/stock/holding/update", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.UpdateStockHoldingsHandler).ServeHTTP)
	// 飞书机器人配置
	privateMux.HandleFunc("POST /v1/stock/notice/config", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.StockNoticeFsSetHandler).ServeHTTP)
	// 测试飞书是否配置成功
	privateMux.HandleFunc("POST /v1/stock/send-msg-test", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SendFsInfoHandler).ServeHTTP)
	// 按指定日期查询历史行情数据
	privateMux.HandleFunc("GET /v1/stock/history/date-range", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockHistoryDataDateRangeHandler).ServeHTTP)

	// 设置apikey
	privateMux.HandleFunc("POST /v1/stock/set-ai-config", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SetAiApiKeyHandler).ServeHTTP)

	// 获取所有apikey
	privateMux.HandleFunc("GET /v1/stock/get-ai-config", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetAiApiKeyHandler).ServeHTTP)

	// 修改补仓信息
	privateMux.HandleFunc("GET /v1/stock/average-down-update", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SetAverageDownInfoHandler).ServeHTTP)

	// 配置监控
	privateMux.HandleFunc("POST /v1/stock/set-alerts", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SetStockTriggeringRulesAlertsHandler).ServeHTTP)

	// 获取配置监控列表
	privateMux.HandleFunc("GET /v1/stock/alerts-list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockTriggeringRulesAlertsHandler).ServeHTTP)

	// stock 信息标记
	privateMux.HandleFunc("POST /v1/stock/set-tag", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SetStockTaggingHandler).ServeHTTP)

	// 获取stock 的标记信息
	privateMux.HandleFunc("GET /v1/stock/get-tag", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockTaggingHandler).ServeHTTP)

	authMux := middleware.AuthMiddleware(privateMux, rdbConn)

	// 本地测试允许跨域
	//corsMux := middleware.AllowCorsMiddleware(authMux)

	// 公共可以访问, 不需要走中间件验证
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("POST /v1/auth", tollbooth.LimitFuncHandler(lmt, verifyApp.VerifyHandler.Auth).ServeHTTP)
	publicMux.HandleFunc("POST /v1/create-category", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.CreateCategoriesHandler).ServeHTTP)
	publicMux.HandleFunc("POST /v1/create-problem", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.CreateProblemsHandler).ServeHTTP)
	publicMux.HandleFunc("GET /v1/get-category", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.GetCategoriesHandler).ServeHTTP)
	publicMux.HandleFunc("GET /v1/get-problem", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.GetProblemsHandler).ServeHTTP)
	publicMux.HandleFunc("POST /v1/del-problem", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.DeleteProblemHandler).ServeHTTP)
	publicMux.HandleFunc("POST /v1/del-category", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.DeleteCategoryHandler).ServeHTTP)
	publicMux.HandleFunc("POST /v1/update-problem-category", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.UpdateProblemCategoryHandler).ServeHTTP)
	publicMux.HandleFunc("POST /v1/upload-file", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.UploadFileHandler).ServeHTTP)

	// 总的路由控制
	rootMux := http.NewServeMux()
	rootMux.Handle("/", authMux)
	rootMux.Handle("/v1/auth", publicMux)
	rootMux.Handle("/v1/create-category", publicMux)
	rootMux.Handle("/v1/create-problem", publicMux)
	rootMux.Handle("/v1/get-category", publicMux)
	rootMux.Handle("/v1/get-problem", publicMux)
	rootMux.Handle("/v1/del-problem", publicMux)
	rootMux.Handle("/v1/del-category", publicMux)
	rootMux.Handle("/v1/update-problem-category", publicMux)
	rootMux.Handle("/v1/upload-file", publicMux)

	log.Println(fmt.Sprintf("Server started on :%d, version: %s", config.HttpPort, config.Version))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.HttpPort), rootMux))
}
