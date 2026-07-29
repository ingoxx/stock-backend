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

func (ds *DocService) CreateCategories(userID uint, data domain.Category) (domain.Category, error) {
	return ds.repo.CreateCategories(userID, data)
}

func (ds *DocService) CreateProblems(userID uint, data domain.Problem) (*domain.Problem, error) {
	return ds.repo.CreateProblems(userID, data)
}

func (ds *DocService) GetProblems(userID uint, categoryID uint, keyword string, page int) ([]*domain.Problem, int64, error) {
	return ds.repo.GetProblems(userID, categoryID, keyword, page)
}

func (ds *DocService) GetCategories(userID uint, page int) ([]domain.Category, int64, error) {
	return ds.repo.GetCategories(userID, page)
}

func (ds *DocService) DeleteCategory(id uint, userID uint) error {
	return ds.repo.DeleteCategory(id, userID)
}

func (ds *DocService) DeleteProblem(id uint, userID uint) error {
	return ds.repo.DeleteProblem(id, userID)
}

func (ds *DocService) UpdateProblemCategory(problemID uint, newCategoryID uint, userID uint) error {
	return ds.repo.UpdateProblemCategory(problemID, newCategoryID, userID)
}

func (ds *DocService) UploadFile(problemID uint, uploaderID uint, fileName string, src io.Reader) (*domain.FileItem, error) {
	return ds.repo.UploadFile(problemID, uploaderID, fileName, src)
}

func (ds *DocService) DeleteFilesByProblemID(problemID, fileID uint) error {
	return ds.repo.DeleteFilesByProblemID(problemID, fileID)
}

func (ds *DocService) LoginUser(username, password string) (*domain.User, error) {
	return ds.repo.LoginUser(username, password)
}

func (ds *DocService) RegisterUser(user *domain.User) (*domain.User, error) {
	return ds.repo.RegisterUser(user)
}

func (ds *DocService) ChangePassword(username string, oldPassword, newPassword string) error {
	return ds.repo.ChangePassword(username, oldPassword, newPassword)
}

func (ds *DocService) ShareCategoryToUsers(categoryID uint, operatorID uint, targetUserIDs []uint) error {
	return ds.repo.ShareCategoryToUsers(categoryID, operatorID, targetUserIDs)
}

func (ds *DocService) ShareProblemToUsers(problemID uint, operatorID uint, targetUserIDs []uint) error {
	return ds.repo.ShareProblemToUsers(problemID, operatorID, targetUserIDs)
}

func (ds *DocService) GetUserList(page int) ([]domain.User, int64, error) {
	return ds.repo.GetUserList(page)
}
