package server

import (
	"github.com/ingoxx/stock-backend/internal/handler"
	"github.com/ingoxx/stock-backend/internal/repository/redis"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/pkg/initial/rds"
)

type GoldenApp struct {
	GoldenHandler *handler.GoldenHandler
}

func NewGoldenApp() *GoldenApp {
	// 1. 初始化 Repos
	goldenRepo := redis.NewGoldRepo(rds.Rds)

	// 2. 初始化 Services
	goldenSvc := service.NewUserService(goldenRepo)

	// 3. 初始化 Handlers
	goldenHandler := handler.NewGoldenHandler(goldenSvc)
	return &GoldenApp{
		GoldenHandler: goldenHandler,
	}
}
