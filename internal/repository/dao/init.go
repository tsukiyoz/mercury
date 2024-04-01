package dao

import (
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&articleDao.Article{},
		&articleDao.PublishedArticle{},
		&AsyncSms{},
		&Task{},
	)
}
