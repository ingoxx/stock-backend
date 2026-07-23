package my

import (
	"github.com/ingoxx/stock-backend/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMy() *gorm.DB {
	dsn := config.MyDsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	return db
	// 2. 自动迁移：如果数据库中没有 users 表，GORM 会自动创建
	// 如果表已经存在，但结构体多了一些新字段，GORM 会自动追加这些新列
	//db.AutoMigrate(&User{})
}
