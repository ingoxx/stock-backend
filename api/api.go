package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/didip/tollbooth"
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/cmd/server"
	"github.com/ingoxx/stock-backend/configs"
	"github.com/ingoxx/stock-backend/internal/middleware"
)

func Start() {
	var rdbConn = make(map[int]*redis.Client)

	goldenApp := server.NewGoldenApp(rdbConn)
	stockApp := server.NewStockApp(rdbConn)
	verifyApp := server.NewVerifyApp(rdbConn)

	lmt := tollbooth.NewLimiter(configs.MaxReqFrequency, nil)

	// 需要走中间件验证
	privateMux := http.NewServeMux()
	privateMux.HandleFunc("/v2/golden/list", tollbooth.LimitFuncHandler(lmt, goldenApp.GoldenHandler.GetGoldenPriceListHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/golden/set", tollbooth.LimitFuncHandler(lmt, goldenApp.GoldenHandler.SetGoldenPriceHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockListHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/days/detail", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockInfoForDataListHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/industry/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockIndustryListHandler).ServeHTTP)
	// 一并返回飞书预警信息是否已经配置
	privateMux.HandleFunc("/v1/stock/industry/up-down", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetIndustryStockUpDownHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/market-data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockMarketDataHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/switch", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockDataSwitchHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/run-status", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockDataStatusHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/industry/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetIndustryDataHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/history/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockCusDaysDataHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/info/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockInfoDataHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/real-time/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockRealTimeDataHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/real-time/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockRealTimeListHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/self-selected-del", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.DelSelfSelectedStockHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/trade-status/update", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.UpdateStockDealStatusHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/trade-history/add", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.AddHistoryTradeDataHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/trade-history/list", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetHistoryTradeDataListHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/real-time/switch", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.StockRealTimeInfoSwitchHandler).ServeHTTP)
	privateMux.HandleFunc("/v2/stock/real-time/data", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockRtDataHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/notice/switch", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.StockNoticeSwitchHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/filter/good", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetGoodStocksHandler).ServeHTTP)
	privateMux.HandleFunc("/v1/stock/filter/good/history", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.FilterGoodStocksHistoryHandler).ServeHTTP)
	// 修改持仓列表中的股票成本以及股数
	privateMux.HandleFunc("POST /v1/stock/holding/update", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.UpdateStockHoldingsHandler).ServeHTTP)
	// 飞书机器人配置
	privateMux.HandleFunc("POST /v1/stock/notice/config", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.StockNoticeFsSetHandler).ServeHTTP)
	// 测试飞书是否配置成功
	privateMux.HandleFunc("POST /v1/stock/send-msg-test", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.SendFsInfoHandler).ServeHTTP)
	// 按指定日期查询历史行情数据
	privateMux.HandleFunc("GET /v1/stock/history/date-range", tollbooth.LimitFuncHandler(lmt, stockApp.StockHandler.GetStockHistoryDataDateRangeHandler).ServeHTTP)
	authMux := middleware.AuthMiddleware(privateMux, rdbConn)

	// 本地测试允许跨域
	//corsMux := middleware.AllowCorsMiddleware(authMux)

	// 公共可以访问, 不需要走中间件验证
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("POST /v1/auth", tollbooth.LimitFuncHandler(lmt, verifyApp.VerifyHandler.Auth).ServeHTTP)

	// 总的路由控制
	rootMux := http.NewServeMux()
	rootMux.Handle("/", authMux)
	rootMux.Handle("/v1/auth", publicMux)

	log.Println(fmt.Sprintf("Server started on :%d, version: %s", configs.HttpPort, configs.Version))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", configs.HttpPort), rootMux))
}
