/**
 * @author tsukiyo
 * @date 2023-08-11 1:29
 */

package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate = errors.New("parameter duplicate")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDao struct {
	db *gorm.DB
}

func (dao *UserDao) Create(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *UserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (dao *UserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return user, err
}

func (dao *UserDao) UpdateById(ctx *gin.Context, uid int64, nickname string, birthday int64, biography string) error {
	return dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"nick_name": nickname,
		"birthday":  birthday,
		"biography": biography,
	}).Error
}

func (dao *UserDao) FindById(ctx *gin.Context, uid int64) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", uid).First(&user).Error
	return user, err
}

type User struct {
	Id        int64          `gorm:"primaryKey,autoIncrement"`
	Birthday  int64          `gorm:"default:0"`
	Email     sql.NullString `gorm:"unique"`
	Phone     sql.NullString `gorm:"unique"`
	NickName  string         `gorm:"default:小书虫"`
	Password  string         `gorm:"not null"`
	Biography string         `gorm:"default:这个用户很懒什么都没有留下"`
	CreateAt  int64
	UpdateAt  int64
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}
