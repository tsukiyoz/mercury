package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

type Interactive struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`

	ReadCnt     int64
	FavoriteCnt int64
	LikeCnt     int64

	Ctime int64
	Utime int64
}

func (i Interactive) ID() int64 {
	return i.Id
}

type Like struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	Uid   int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	// 0-unlike, 1-like
	Status uint8
	Ctime  int64
	Utime  int64
}

type Favorites struct {
	Id   int64  `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"type=varchar(1024)"`
	Uid  int64

	Ctime int64
	Utime int64
}

type FavoriteItem struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	Uid   int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	// Favorites ID
	Fid   int64 `gorm:"index"`
	Ctime int64
	Utime int64
}

//go:generate mockgen -source=./interactive.go -package=daomocks -destination=mocks/interactive.mock.go InteractiveDAO
type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (Like, error)
	DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	InsertFavoriteItem(ctx context.Context, ci FavoriteItem) error
	GetFavoriteInfo(ctx context.Context, biz string, bizId, uid int64) (FavoriteItem, error)
	BatchIncrReadCnt(ctx context.Context, biz string, ids []int64) error
	GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error)
}

var _ InteractiveDAO = (*GORMInteractiveDAO)(nil)

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}

func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return dao.incrReadCnt(dao.db.WithContext(ctx), biz, bizId)
}

func (dao *GORMInteractiveDAO) incrReadCnt(tx *gorm.DB, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return tx.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"read_cnt": gorm.Expr("`read_cnt`+1"),
			"utime":    now,
		}),
	}).Create(&Interactive{
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
		Biz:     biz,
		BizId:   bizId,
	}).Error
}

func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"status": 1,
				"utime":  now,
			}),
		}).Create(&Like{
			Uid:    uid,
			BizId:  bizId,
			Biz:    biz,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("`like_cnt` + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			LikeCnt: 1,
			Biz:     biz,
			BizId:   bizId,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
	return err
}

func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (Like, error) {
	var like Like
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).
		First(&like).Error
	return like, err
}

func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&Like{}).
			Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).
			Updates(map[string]any{
				"status": 0,
				"utime":  now,
			}).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("`like_cnt` - 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   bizId,
			LikeCnt: 0,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
	return err
}

func (dao *GORMInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var intr Interactive
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ?", biz, bizId).
		First(&intr).Error
	return intr, err
}

func (dao *GORMInteractiveDAO) InsertFavoriteItem(ctx context.Context, ci FavoriteItem) error {
	now := time.Now().UnixMilli()
	ci.Ctime, ci.Utime = now, now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&ci).Error
		if err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"favorite_cnt": gorm.Expr("`favorite_cnt` + 1"),
				"utime":        now,
			}),
		}).Create(&Interactive{
			Biz:         ci.Biz,
			BizId:       ci.BizId,
			FavoriteCnt: 1,
			Ctime:       now,
			Utime:       now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) GetFavoriteInfo(ctx context.Context, biz string, bizId, uid int64) (FavoriteItem, error) {
	var favoriteItem FavoriteItem
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).
		First(&favoriteItem).Error
	return favoriteItem, err
}

func (dao *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, biz string, ids []int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(ids); i++ {
			err := dao.incrReadCnt(tx, biz, ids[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *GORMInteractiveDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	var intrs []Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? AND id IN ?", biz, ids).Find(&intrs).Error
	return intrs, err
}
