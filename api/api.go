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

	// -------------------------------------------------------------
	// 1. 原有的私有业务接口 (走原来的 AuthMiddleware)
	// -------------------------------------------------------------
	privateMux := http.NewServeMux()

	// gold变动接口
	privateMux.HandleFunc("GET /v2/golden/list", tollbooth.LimitFuncHandler(lmt, goldenApp.GoldenHandler.GetGoldenPriceListHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/golden/set", tollbooth.LimitFuncHandler(lmt, goldenApp.GoldenHandler.SetGoldenPriceHandler).ServeHTTP)

	// stock接口
	privateMux.HandleFunc("GET /v1/stock/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockListHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/days/detail", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockInfoForDataListHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/industry/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockIndustryListHandler).ServeHTTP)
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
	privateMux.HandleFunc("GET /v1/stock/filter/good/history", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.FilterGoodStocksHistoryHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/self-selected/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetSelfSelectedStockListHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/self-selected/add", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.AddSelfSelectedStockHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v2/stock/self-selected/del", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SelfSelectedStockDelHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/sh-index", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetShIndexRealTimeDataHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/capital-inflow", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetCapitalInflowDataHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/holding/update", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.UpdateStockHoldingsHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/notice/config", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.StockNoticeFsSetHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/send-msg-test", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SendFsInfoHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/history/date-range", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockHistoryDataDateRangeHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/set-ai-config", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SetAiApiKeyHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/get-ai-config", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetAiApiKeyHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/average-down-update", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SetAverageDownInfoHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/set-alerts", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SetStockTriggeringRulesAlertsHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/alerts-list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockTriggeringRulesAlertsHandler).ServeHTTP)
	privateMux.HandleFunc("POST /v1/stock/set-tag", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SetStockTaggingHandler).ServeHTTP)
	privateMux.HandleFunc("GET /v1/stock/get-tag", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockTaggingHandler).ServeHTTP)

	// 原有的 Auth 中间件包裹
	authMux := middleware.AuthMiddleware(privateMux, rdbConn)

	// -------------------------------------------------------------
	// 2. 公共开放接口 (不需要 JWT 验证，无需任何 Auth 中间件)
	// -------------------------------------------------------------
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("POST /v1/auth", tollbooth.LimitFuncHandler(lmt, verifyApp.VerifyHandler.Auth).ServeHTTP)
	// 文档系统的注册与登录接口（公开）
	publicMux.HandleFunc("POST /v1/doc/register", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.RegisterHandler).ServeHTTP)
	publicMux.HandleFunc("POST /v1/doc/login", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.LoginHandler).ServeHTTP)
	publicMux.HandleFunc("POST /v1/doc/change-password", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.ChangePasswordHandler).ServeHTTP)

	// -------------------------------------------------------------
	// 3. 文档业务接口 (独占走 JWTAuthMiddleware 验证)
	// -------------------------------------------------------------
	docMux := http.NewServeMux()
	// 修改点：注册到 docMux 上，而不是 publicMux 上
	docMux.HandleFunc("POST /v1/create-category", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.CreateCategoriesHandler).ServeHTTP)
	docMux.HandleFunc("POST /v1/create-problem", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.CreateProblemsHandler).ServeHTTP)
	docMux.HandleFunc("GET /v1/get-category", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.GetCategoriesHandler).ServeHTTP)
	docMux.HandleFunc("GET /v1/get-problem", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.GetProblemsHandler).ServeHTTP)
	docMux.HandleFunc("POST /v1/del-problem", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.DeleteProblemHandler).ServeHTTP)
	docMux.HandleFunc("POST /v1/del-category", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.DeleteCategoryHandler).ServeHTTP)
	docMux.HandleFunc("POST /v1/update-problem-category", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.UpdateProblemCategoryHandler).ServeHTTP)
	docMux.HandleFunc("POST /v1/upload-file", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.UploadFileHandler).ServeHTTP)
	docMux.HandleFunc("POST /v1/del-file", tollbooth.LimitFuncHandler(lmt, docApp.DocHandler.DeleteFileHandler).ServeHTTP)

	// 用 JWTAuthMiddleware 仅包裹 docMux
	docAuth := middleware.JWTAuthMiddleware(docMux)

	// -------------------------------------------------------------
	// 4. 总路由控制挂载
	// -------------------------------------------------------------
	rootMux := http.NewServeMux()

	// 挂载公共接口 (无需验证)
	rootMux.Handle("/v1/auth", publicMux)
	rootMux.Handle("/v1/doc/register", publicMux)
	rootMux.Handle("/v1/doc/login", publicMux)
	rootMux.Handle("/v1/doc/change-password", publicMux)

	// 修改点：挂载文档接口，使用经过 JWT 验证包裹的 docAuth
	rootMux.Handle("/v1/create-category", docAuth)
	rootMux.Handle("/v1/create-problem", docAuth)
	rootMux.Handle("/v1/get-category", docAuth)
	rootMux.Handle("/v1/get-problem", docAuth)
	rootMux.Handle("/v1/del-problem", docAuth)
	rootMux.Handle("/v1/del-category", docAuth)
	rootMux.Handle("/v1/update-problem-category", docAuth)
	rootMux.Handle("/v1/upload-file", docAuth)
	rootMux.Handle("/v1/del-file", docAuth)

	// 兜底挂载：其他所有黄金/股票接口走原有的 authMux
	rootMux.Handle("/", authMux)

	log.Println(fmt.Sprintf("Server started on :%d, version: %s", config.HttpPort, config.Version))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.HttpPort), rootMux))
}
