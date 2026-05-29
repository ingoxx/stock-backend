package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/handler"
	rdbRepo "github.com/ingoxx/stock-backend/internal/repository/redis"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/pkg/initial/rds"
)

var (
	db = 10
)

type VerifyApp struct {
	VerifyHandler *handler.VerifyHandler
	VerifyService *service.VerifyService
}

type VerifyResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewVerifyApp(rc map[int]*redis.Client) *VerifyApp {
	validate := validator.New()
	client := rds.GetRedisClient(db, rc)
	verifyRepo := rdbRepo.NewVerifyRepo(client)
	verifyService := service.NewVerifyService(verifyRepo)
	verifyHandler := handler.NewVerifyHandler(verifyService, validate)

	return &VerifyApp{
		VerifyHandler: verifyHandler,
		VerifyService: verifyService,
	}
}
