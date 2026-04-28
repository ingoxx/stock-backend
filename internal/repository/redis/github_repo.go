package redis

import (
	"sync"

	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/domain"
)

type GithubCallBackRepo struct {
	mu     sync.RWMutex
	client *redis.Client
	wg     sync.WaitGroup
}

func NewGithubCallBackRepo(client *redis.Client) domain.GithubCallbackRepository {
	return &GithubCallBackRepo{
		client: client,
	}
}

func (repo *GithubCallBackRepo) GitHubCallBackApi() error {
	return nil
}
