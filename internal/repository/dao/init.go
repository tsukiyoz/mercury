/**
 * @author tsukiyo
 * @date 2023-08-11 21:58
 */

package dao

import (
	"github.com/tsukaychan/webook/internal/repository/dao/article"
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &article.Article{}, &article.PublishedArticle{})
}
