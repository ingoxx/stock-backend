package service

import "github.com/ingoxx/stock-backend/internal/domain"

type GithubCallBackService struct {
	repo domain.GithubCallbackRepository
}

func NewGithubCallBackService(repo domain.GithubCallbackRepository) *GithubCallBackService {
	return &GithubCallBackService{repo: repo}
}

func (g *GithubCallBackService) GitHubCallBackApi() error {
	return g.repo.GitHubCallBackApi()
}
