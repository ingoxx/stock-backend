package server

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/stock-backend/internal/domain"
	"github.com/ingoxx/stock-backend/internal/handler"
	"github.com/ingoxx/stock-backend/internal/repository/mysql"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/pkg/initial/my"
	"github.com/ingoxx/stock-backend/utils"
	"gorm.io/gorm"
)

type DocApp struct {
	DocHandler *handler.DocHandler
}

func NewDocApp() *DocApp {
	validate := validator.New()
	gd := my.InitMy()

	// 1. 优先创建 users 表（如果不存在）
	if err := gd.AutoMigrate(&domain.User{}); err != nil {
		panic(fmt.Sprintf("迁移 users 表失败: %s", err))
	}

	// 2. 检查或创建初始默认管理员用户（ID=1），作为历史老数据的归属者
	var defaultAdmin domain.User
	err := gd.First(&defaultAdmin, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hashedPassword, _ := utils.HashPassword("admin123") // 默认初始密码 admin123
		defaultAdmin = domain.User{
			ID:       1,
			Username: "admin",
			Password: hashedPassword,
		}
		if err := gd.Create(&defaultAdmin).Error; err != nil {
			panic(fmt.Sprintf("创建默认管理员账号失败: %s", err))
		}
	}

	// 3. 执行现有表结构自动升级 (加字段、加索引、自动建多对多表 problem_editors)
	err = gd.AutoMigrate(
		&domain.Category{},
		&domain.Problem{},
		&domain.FileItem{},
	)
	if err != nil {
		panic(fmt.Sprintf("数据表结构升级失败: %s", err))
	}

	// 4. 数据洗牌（Backfill）：确保历史存量数据中 null 或 0 的外键均绑定到默认管理员 (ID=1)
	gd.Model(&domain.Category{}).
		Where("creator_id = 0 OR creator_id IS NULL").
		Updates(map[string]interface{}{
			"creator_id":    defaultAdmin.ID,
			"updated_by_id": defaultAdmin.ID,
		})

	gd.Model(&domain.Problem{}).
		Where("creator_id = 0 OR creator_id IS NULL").
		Updates(map[string]interface{}{
			"creator_id":    defaultAdmin.ID,
			"updated_by_id": defaultAdmin.ID,
		})

	gd.Model(&domain.FileItem{}).
		Where("uploader_id = 0 OR uploader_id IS NULL").
		Updates(map[string]interface{}{
			"uploader_id": defaultAdmin.ID,
		})

	docRepo := mysql.NewDocRepo(gd)
	docSvc := service.NewDocService(docRepo)
	docHandler := handler.NewDocHandler(docSvc, validate)

	return &DocApp{
		DocHandler: docHandler,
	}
}
