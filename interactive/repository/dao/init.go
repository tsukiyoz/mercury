package dao

import (
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&Interactive{},
		// user likes
		&Like{},
		// user favorites
		&Favorites{},
		// favorite item relationship with favorites
		&FavoriteItem{},
	)
}
