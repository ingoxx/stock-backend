package api

import (
	"github.com/ingoxx/stock-backend/cmd/server"
	"log"
	"net/http"
)

func Start() {
	app := server.NewGoldenApp()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/golden/list", app.GoldenHandler.GetGoldenPriceListHandler)
	mux.HandleFunc("/v1/golden/set", app.GoldenHandler.SetGoldenPriceHandler)
	log.Println("Server started on :11806")
	log.Fatal(http.ListenAndServe(":11806", mux))
}
