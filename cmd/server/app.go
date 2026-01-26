package server

import (
	"github.com/ingoxx/stock-backend/internal/handler"
	rdbRepo "github.com/ingoxx/stock-backend/internal/repository/redis"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/redis/go-redis"
)

type GoldenApp struct {
	GoldenHandler *handler.GoldenHandler
}

func NewGoldenApp(rc *redis.Client) *GoldenApp {
	//rdb, err := rds.NewRedisClient()
	//if err != nil {
	//	panic(err)
	//}

	goldenRepo := rdbRepo.NewGoldRepo(rc)

	goldenSvc := service.NewUserService(goldenRepo)

	goldenHandler := handler.NewGoldenHandler(goldenSvc)

	return &GoldenApp{
		GoldenHandler: goldenHandler,
	}
}
