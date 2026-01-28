package api

import (
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/cmd/server"
	"github.com/ingoxx/stock-backend/configs"
	"github.com/ingoxx/stock-backend/internal/middleware"
	"log"
	"net/http"
)

func Start() {
	var rdbConn = make(map[int]*redis.Client)

	app := server.NewGoldenApp(rdbConn)

	lmt := tollbooth.NewLimiter(configs.MaxReqFrequency, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/golden/list", tollbooth.LimitFuncHandler(lmt, app.GoldenHandler.GetGoldenPriceListHandler).ServeHTTP)
	mux.HandleFunc("/v1/golden/set", tollbooth.LimitFuncHandler(lmt, app.GoldenHandler.SetGoldenPriceHandler).ServeHTTP)

	authMux := middleware.AuthMiddleware(mux, rdbConn)
	//corsMux := middleware.AllowCorsMiddleware(authMux)

	log.Println(fmt.Sprintf("Server started on :%d, version: %s", configs.HttpPort, configs.Version))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", configs.HttpPort), authMux))
}
