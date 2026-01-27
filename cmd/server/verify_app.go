package server

import (
	"github.com/go-redis/redis"
	rdbRepo "github.com/ingoxx/stock-backend/internal/repository/redis"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/pkg/initial/rds"
)

var (
	db = 10
)

type VerifyApp struct {
	VerifyService *service.VerifyService
}

type VerifyResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewVerifyApp(rc map[int]*redis.Client) *VerifyApp {
	var client = rds.GetRedisClient(db, rc)
	repo := rdbRepo.NewVerifyRepo(client)
	verifyService := service.NewVerifyService(repo)

	return &VerifyApp{
		VerifyService: verifyService,
	}
}
