package main

import (
	"github.com/ingoxx/stock-backend/api"
	"log"
)

func main() {
	log.Println("The HTTP server started successfully.")

	api.Start()
}
