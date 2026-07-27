package domain

import (
	"io"
	"time"
)

// User 用户模型 (映射表名: users)
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username" validate:"required" gorm:"type:varchar(50);unique;not null;index"`

	// json:"-" 避免密码在序列化为 JSON 响应时泄露
	Password string `json:"-" validate:"required" gorm:"type:varchar(255);not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Category struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string `json:"name" validate:"required" gorm:"type:varchar(100);not null"`
	IsShared bool   `json:"is_shared" gorm:"default:false;index"`

	// 补上 default:1 容错，旧数据增加此列时会自动填 1
	CreatorID   uint `json:"creator_id" gorm:"default:1;not null;index"`
	UpdatedByID uint `json:"updated_by_id" gorm:"default:1;not null;index"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Problems  []Problem `json:"problems,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
	Creator   *User     `json:"creator,omitempty" gorm:"foreignKey:CreatorID;references:ID"`
	UpdatedBy *User     `json:"updated_by,omitempty" gorm:"foreignKey:UpdatedByID;references:ID"`
}

type Problem struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	CategoryID uint   `json:"category_id" validate:"required" gorm:"not null;index"`
	Title      string `json:"title" validate:"required" gorm:"type:varchar(255);not null"`
	Solution   string `json:"solution" validate:"required" gorm:"type:text"`
	IsShared   bool   `json:"is_shared" gorm:"default:false;index"`

	// 补上 default:1 容错
	CreatorID   uint `json:"creator_id" gorm:"default:1;not null;index"`
	UpdatedByID uint `json:"updated_by_id" gorm:"default:1;not null;index"`
	Version     int  `json:"version" gorm:"default:1"`

	Files   []FileItem `json:"file_url,omitempty" gorm:"foreignKey:ProblemID;references:ID"`
	Editors []User     `json:"editors,omitempty" gorm:"many2many:problem_editors;"`

	CreatedAt time.Time `json:"date" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Category  *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
	Creator   *User     `json:"creator,omitempty" gorm:"foreignKey:CreatorID;references:ID"`
	UpdatedBy *User     `json:"updated_by,omitempty" gorm:"foreignKey:UpdatedByID;references:ID"`
}

type FileItem struct {
	ID        uint `json:"id" gorm:"primaryKey;autoIncrement"`
	ProblemID uint `json:"problem_id" gorm:"not null;index"`
	// 补上 default:1 容错
	UploaderID uint      `json:"uploader_id" gorm:"default:1;not null;index"`
	Name       string    `json:"name" gorm:"type:varchar(100)"`
	URL        string    `json:"url" gorm:"type:varchar(255);not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`

	Uploader *User `json:"uploader,omitempty" gorm:"foreignKey:UploaderID;references:ID"`
}

type DocRepository interface {
	GetCategories(userID uint, page int) ([]Category, int64, error)
	GetProblems(userID uint, page int) ([]*Problem, int64, error)
	CreateCategories(userID uint, data Category) (Category, error)
	CreateProblems(userID uint, data Problem) (*Problem, error)
	DeleteCategory(id uint, userID uint) error
	DeleteProblem(id uint, userID uint) error
	UpdateProblemCategory(problemID uint, newCategoryID uint, userID uint) error
	UploadFile(problemID uint, uploaderID uint, fileName string, src io.Reader) (*FileItem, error)
	DeleteFilesByProblemID(problemID uint) error
	RegisterUser(user *User) (*User, error)
	LoginUser(username, password string) (*User, error)
}
