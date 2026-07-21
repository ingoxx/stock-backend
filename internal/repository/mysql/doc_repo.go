package mysql

import (
	"sync"
	"time"

	"github.com/ingoxx/stock-backend/internal/domain"
	"gorm.io/gorm"
)

const (
	pageSize = 10
)

type DocRepo struct {
	db *gorm.DB
	mu sync.RWMutex
}

func NewDocRepo(db *gorm.DB) domain.DocRepository {
	return &DocRepo{
		db: db,
	}
}

func (dr *DocRepo) CreateCategories(data domain.Categories) ([]domain.Categories, error) {
	if err := dr.db.Create(&data).Error; err != nil {
		return nil, err
	}

	var ds []domain.Categories
	if err := dr.db.Find(&ds).Error; err != nil {
		return nil, err
	}

	return ds, nil
}

func (dr *DocRepo) CreateProblems(data domain.Problems) ([]*domain.Problems, error) {
	data.Date = time.Now().Format("2006-01-02 15:04:05")

	if err := dr.db.Create(&data).Error; err != nil {
		return nil, err
	}

	var dp []*domain.Problems
	if err := dr.db.Find(&dp).Error; err != nil {
		return nil, err
	}

	return dp, nil
}
func (dr *DocRepo) GetCategories() ([]domain.Categories, error) {
	return nil, nil
}
func (dr *DocRepo) GetProblems() ([]*domain.Problems, error) {
	return nil, nil
}
