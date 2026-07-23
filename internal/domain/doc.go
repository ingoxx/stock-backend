package domain

import (
	"io"
	"time"
)

// Category 分类模型
type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" validate:"required" gorm:"type:varchar(100);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoCreateTime"`
	Problems  []Problem `json:"problems,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
}

// Problem 问题模型
type Problem struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	CategoryID uint   `json:"category_id" validate:"required" gorm:"not null;index"`
	Title      string `json:"title" validate:"required" gorm:"type:varchar(255);not null"`
	Solution   string `json:"solution" validate:"required" gorm:"type:text"`

	// 一对多关联：一个问题包含多个文件附件
	// foreignKey:ProblemID 表示 FileItem 表中通过 ProblemID 外键关联到这里的 ID
	Files []FileItem `json:"file_url,omitempty" gorm:"foreignKey:ProblemID;references:ID"`

	CreatedAt time.Time `json:"date" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoCreateTime"`

	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
}

// FileItem 附件文件模型 (自动映射表名: file_items)
type FileItem struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	ProblemID uint   `json:"problem_id" gorm:"not null;index"` // 外键，建立索引
	Name      string `json:"name" gorm:"type:varchar(100)"`
	URL       string `json:"url" gorm:"type:varchar(255);not null"`
}

type DocRepository interface {
	GetCategories(page int) ([]Category, error)
	GetProblems(page int) ([]*Problem, error)
	CreateCategories(data Category) (Category, error)
	CreateProblems(data Problem) (*Problem, error)
	DeleteCategory(id uint) error
	DeleteProblem(id uint) error
	UpdateProblemCategory(problemID uint, newCategoryID uint) error
	UploadFile(problemID uint, fileName string, src io.Reader) (*FileItem, error)
}
