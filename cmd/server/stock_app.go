package server

import (
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/handler"
	rdbRepo "github.com/ingoxx/stock-backend/internal/repository/redis"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/pkg/initial/rds"
)

type StockApp struct {
	StockHandler *handler.StockHandler
}

func NewStockApp(rc map[int]*redis.Client) *StockApp {
	var db = 11
	var client = rds.GetRedisClient(db, rc)
	stockRepo := rdbRepo.NewStockRepo(client)
	stockSvc := service.NewStockService(stockRepo)
	stockHandler := handler.NewStockHandler(stockSvc)

	return &StockApp{
		StockHandler: stockHandler,
	}
}
