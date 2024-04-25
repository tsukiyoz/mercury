package dao

import (
	articleDao "github.com/tsukaychan/mercury/article/repository/dao"
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&articleDao.Article{},
		&articleDao.PublishedArticle{},
		&AsyncSms{},
		&Task{},
	)
}
