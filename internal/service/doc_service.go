package service

import (
	"io"

	"github.com/ingoxx/stock-backend/internal/domain"
)

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

func (ds *DocService) DeleteCategory(id uint) error {
	return ds.repo.DeleteCategory(id)
}

func (ds *DocService) DeleteProblem(id uint) error {
	return ds.repo.DeleteProblem(id)
}

func (ds *DocService) UpdateProblemCategory(problemID uint, newCategoryID uint) error {
	return ds.repo.UpdateProblemCategory(problemID, newCategoryID)
}

func (ds *DocService) UploadFile(problemID uint, fileName string, src io.Reader) (*domain.FileItem, error) {
	return ds.repo.UploadFile(problemID, fileName, src)
}
