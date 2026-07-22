package server

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/stock-backend/internal/domain"
	"github.com/ingoxx/stock-backend/internal/handler"
	"github.com/ingoxx/stock-backend/internal/repository/mysql"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/pkg/initial/my"
)

type DocApp struct {
	DocHandler *handler.DocHandler
}

func NewDocApp() *DocApp {
	validate := validator.New()
	gd := my.InitMy()

	if err := gd.AutoMigrate(&domain.Category{}); err != nil {
		panic(fmt.Errorf("failed to auto migrate category: %v", err))
	}

	if err := gd.AutoMigrate(&domain.Problem{}); err != nil {
		panic(fmt.Errorf("failed to auto migrate problem: %v", err))
	}

	if err := gd.AutoMigrate(&domain.FileItem{}); err != nil {
		panic(fmt.Errorf("failed to auto migrate file_item: %v", err))
	}

	docRepo := mysql.NewDocRepo(gd)
	docSvc := service.NewDocService(docRepo)
	docHandler := handler.NewDocHandler(docSvc, validate)

	return &DocApp{
		DocHandler: docHandler,
	}
}
