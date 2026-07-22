package mysql

import (
	"github.com/ingoxx/stock-backend/internal/domain"
	"gorm.io/gorm"
)

const (
	pageSize = 10
	saveDir  = "/tmp"
)

type DocRepo struct {
	db *gorm.DB
}

func NewDocRepo(db *gorm.DB) domain.DocRepository {
	return &DocRepo{
		db: db,
	}
}

// CreateCategories 创建分类
func (dr *DocRepo) CreateCategories(data domain.Category) (domain.Category, error) {
	if err := dr.db.Create(&data).Error; err != nil {
		return domain.Category{}, err
	}
	return data, nil
}

// CreateProblems 创建问题
func (dr *DocRepo) CreateProblems(data domain.Problem) (*domain.Problem, error) {
	if err := dr.db.Create(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

// GetProblems 获取问题列表（带分页 + 关联分类 + 关联文件列表）
func (dr *DocRepo) GetProblems(page int) ([]*domain.Problem, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var dp []*domain.Problem
	err := dr.db.
		Preload("Category"). // 预加载分类
		Preload("Files"). // 预加载当前问题的附件列表
		Limit(pageSize).
		Offset(offset).
		Find(&dp).Error

	if err != nil {
		return nil, err
	}

	return dp, nil
}

// GetCategories 获取分类列表（带分页 + 关联问题 + 关联问题的附件）
func (dr *DocRepo) GetCategories(page int) ([]domain.Category, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var ds []domain.Category
	err := dr.db.
		Preload("Problems"). // 预加载分类下的问题
		Preload("Problems.Files"). // 嵌套预加载：同时把问题的附件列表也查出来
		Limit(pageSize).
		Offset(offset).
		Find(&ds).Error

	if err != nil {
		return nil, err
	}

	return ds, nil
}
