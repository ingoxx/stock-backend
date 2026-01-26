package api

import (
	"github.com/ingoxx/stock-backend/cmd/server"
	"github.com/ingoxx/stock-backend/internal/middleware"
	"github.com/ingoxx/stock-backend/pkg/initial/rds"
	"log"
	"net/http"
)

func Start() {
	rdb, err := rds.NewRedisClient()
	if err != nil {
		panic(err)
	}

	app := server.NewGoldenApp(rdb)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/golden/list", app.GoldenHandler.GetGoldenPriceListHandler)
	mux.HandleFunc("/v1/golden/set", app.GoldenHandler.SetGoldenPriceHandler)

	newMux := middleware.AuthMiddleware(mux)

	log.Println("Server started on :11806")

	log.Fatal(http.ListenAndServe(":11806", newMux))
}
