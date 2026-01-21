package api

import (
	"github.com/ingoxx/stock-backend/cmd/server"
	"log"
	"net/http"
)

func Start() {
	mux := http.NewServeMux()
	app := server.NewGoldenApp()

	mux.HandleFunc("/v1/golden/list", app.GoldenHandler.GetGoldenPriceHandler)
	mux.HandleFunc("/v1/golden/set", app.GoldenHandler.SetGoldenPriceHandler)
	log.Println("Server started on :11806")
	log.Fatal(http.ListenAndServe(":11806", mux))
}
