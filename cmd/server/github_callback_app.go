package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/handler"
	rdbRepo "github.com/ingoxx/stock-backend/internal/repository/redis"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/pkg/initial/rds"
)

type GithubApp struct {
	GithubHandler *handler.GithubCallBackHandler
}

func NewGithubApp(rc map[int]*redis.Client) *GithubApp {
	var db = 11
	var client = rds.GetRedisClient(db, rc)
	validate := validator.New()
	githubRepo := rdbRepo.NewGithubCallBackRepo(client)
	githubSvc := service.NewGithubCallBackService(githubRepo)
	githubHandler := handler.NewGithubCallBackHandler(githubSvc, validate)

	return &GithubApp{
		GithubHandler: githubHandler,
	}
}
