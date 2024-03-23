/**
 * @author tsukiyo
 * @date 2023-08-11 21:58
 */

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
		&Interactive{},
		// user likes
		&Like{},
		// user favorites
		&Favorites{},
		// favorite item relationship with favorites
		&FavoriteItem{},
		&Task{},
	)
}
