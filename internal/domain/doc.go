package domain

type Categories struct {
	Id   int    `json:"id"`
	Name string `json:"name"  validate:"required"`
}

type Problems struct {
	Id         int    `json:"id"  validate:"required"`
	CategoryId int    `json:"category_id"  validate:"required"`
	Title      string `json:"title"  validate:"required"`
	Solution   string `json:"solution"  validate:"required"`
	Date       string `json:"date"`
}

type DocRepository interface {
	GetCategories() ([]Categories, error)
	GetProblems() ([]*Problems, error)
	CreateCategories(Categories) ([]Categories, error)
	CreateProblems(Problems) ([]*Problems, error)
}
