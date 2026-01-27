package server

import (
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/handler"
	rdbRepo "github.com/ingoxx/stock-backend/internal/repository/redis"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/pkg/initial/rds"
)

type GoldenApp struct {
	GoldenHandler *handler.GoldenHandler
}

func NewGoldenApp(rc map[int]*redis.Client) *GoldenApp {
	var db = 11
	var client = rds.GetRedisClient(db, rc)

	goldenRepo := rdbRepo.NewGoldRepo(client)

	goldenSvc := service.NewUserService(goldenRepo)

	goldenHandler := handler.NewGoldenHandler(goldenSvc)

	return &GoldenApp{
		GoldenHandler: goldenHandler,
	}
}
