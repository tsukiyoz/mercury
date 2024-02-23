package dao

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/sqlx"
	"github.com/tsukaychan/webook/internal/service/sms"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

type AsyncSmsDao interface {
	Insert(ctx context.Context, s AsyncSms) error
	GetWaitingSMS(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

type AsyncSms struct {
	Id       int64
	Config   sqlx.JsonColumn[SMSConfig]
	RetryCnt int
	RetryMax int
	Status   uint8
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

type SMSConfig struct {
	TplId  string
	Args   []sms.ArgVal
	Phones []string
}

const (
	asyncStatusWaiting = iota
	asyncStatusFailed
	asyncStatusSuccess
)

var _ AsyncSmsDao = (*GORMAsyncSmsDao)(nil)

type GORMAsyncSmsDao struct {
	db *gorm.DB
}

func NewGORMAsyncSmsDao(db *gorm.DB) AsyncSmsDao {
	return &GORMAsyncSmsDao{
		db: db,
	}
}

func (dao *GORMAsyncSmsDao) Insert(ctx context.Context, s AsyncSms) error {
	return dao.db.WithContext(ctx).Create(&s).Error
}

func (dao *GORMAsyncSmsDao) GetWaitingSMS(ctx context.Context) (AsyncSms, error) {
	var as AsyncSms
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("utime < ? and status = ?", endTime, asyncStatusWaiting).First(&as).Error
		if err != nil {
			return err
		}

		err = tx.Model(&AsyncSms{}).Where("id = ?", as.Id).Updates(map[string]any{
			"retry_cnt": gorm.Expr("`retry_cnt` + 1"),
			"utime":     now,
		}).Error
		return err
	})
	return as, err
}

func (dao *GORMAsyncSmsDao) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.Where(ctx).Model(&AsyncSms{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error
}

func (dao *GORMAsyncSmsDao) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id = ? and `retry_cnt`>=`retry_max`", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusFailed,
		}).Error
}
