package service

import "github.com/ingoxx/stock-backend/internal/domain"

type DocService struct {
	repo domain.DocRepository
}

func NewDocService(repo domain.DocRepository) *DocService {
	return &DocService{repo: repo}
}

func (ds *DocService) CreateCategories(data domain.Category) (domain.Category, error) {
	return ds.repo.CreateCategories(data)
}

func (ds *DocService) CreateProblems(data domain.Problem) (*domain.Problem, error) {
	return ds.repo.CreateProblems(data)
}

func (ds *DocService) GetProblems(page int) ([]*domain.Problem, error) {
	return ds.repo.GetProblems(page)
}

func (ds *DocService) GetCategories(page int) ([]domain.Category, error) {
	return ds.repo.GetCategories(page)
}
