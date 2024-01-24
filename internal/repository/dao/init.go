/**
 * @author tsukiyo
 * @date 2023-08-11 21:58
 */

package dao

import "gorm.io/gorm"

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Article{})
}
