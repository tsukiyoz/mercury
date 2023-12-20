/**
 * @author tsukiyo
 * @date 2023-08-11 1:29
 */

package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate = errors.New("parameter duplicate")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

var _ UserDao = (*UserGormDao)(nil)

type UserDao interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindById(ctx context.Context, uid int64) (User, error)
	UpdateNonZeroFields(ctx context.Context, user User) error
	FindByWechat(ctx context.Context, openID string) (User, error)
}

type UserGormDao struct {
	db *gorm.DB
}

func (dao *UserGormDao) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.CreateAt = now
	u.UpdateAt = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *UserGormDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (dao *UserGormDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return user, err
}

func (dao *UserGormDao) FindById(ctx context.Context, uid int64) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", uid).First(&user).Error
	return user, err
}

func (dao *UserGormDao) UpdateNonZeroFields(ctx context.Context, user User) error {
	return dao.db.WithContext(ctx).Updates(&user).Error
}

func (dao *UserGormDao) FindByWechat(ctx context.Context, openID string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openID).First(&u).Error
	return u, err
}

type User struct {
	Id            int64 `gorm:"primaryKey,autoIncrement"`
	Birthday      sql.NullInt64
	Email         sql.NullString `gorm:"unique"`
	Phone         sql.NullString `gorm:"unique"`
	NickName      sql.NullString
	Password      string         `gorm:"not null"`
	AboutMe       sql.NullString `gorm:"default:这个用户很懒什么都没有留下;type=varchar(1024)"`
	WechatUnionID sql.NullString `gorm:"unique"`
	WechatOpenID  sql.NullString `gorm:"unique"`
	CreateAt      int64
	UpdateAt      int64
}

func NewUserGormDao(db *gorm.DB) UserDao {
	return &UserGormDao{
		db: db,
	}
}
