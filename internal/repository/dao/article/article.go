package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, article PublishedArticle) error
	SyncStatus(ctx context.Context, atcl Article) error
}

// Article Production Library
type Article struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Title    string `gorm:"type=varchar(1024)"`
	Content  string `gorm:"blob"`
	AuthorId int64  `gorm:"index=aid_ctime"`
	Status   uint8
	Ctime    int64 `gorm:"index=aid_ctime"`
	Utime    int64
}

// PublishedArticle OnLive Library
type PublishedArticle struct {
	Article
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, atcl Article) (int64, error) {
	now := time.Now().UnixMilli()
	atcl.Ctime = now
	atcl.Utime = now
	err := dao.db.WithContext(ctx).Create(&atcl).Error
	return atcl.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, atcl Article) error {
	now := time.Now().UnixMilli()
	atcl.Utime = now
	res := dao.db.WithContext(ctx).Model(&atcl).
		Where("id=? AND author_id=?", atcl.Id, atcl.AuthorId).
		Updates(map[string]any{
			"title":   atcl.Title,
			"content": atcl.Content,
			"status":  atcl.Status,
			"utime":   atcl.Utime,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("update atcl failed, perhaps invalid id: id %d author_id %d", atcl.Id, atcl.AuthorId)
	}
	return nil
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, atcl Article) (int64, error) {
	var id = atcl.Id
	err := dao.db.Transaction(func(tx *gorm.DB) (err error) {
		txDao := NewGORMArticleDAO(tx)

		if atcl.Id > 0 {
			err = txDao.UpdateById(ctx, atcl)
		} else {
			id, err = txDao.Insert(ctx, atcl)
		}
		if err != nil {
			return err
		}

		atcl.Id = id
		return txDao.Upsert(ctx, PublishedArticle{atcl})
	})
	return id, err
}

func (dao *GORMArticleDAO) Upsert(ctx context.Context, atcl PublishedArticle) error {
	now := time.Now().UnixMilli()
	atcl.Utime, atcl.Ctime = now, now
	err := dao.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   atcl.Title,
			"content": atcl.Content,
			"status":  atcl.Status,
			"utime":   atcl.Utime,
		}),
	}).Create(&atcl).Error
	return err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, atcl Article) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", atcl.Id, atcl.AuthorId).Updates(map[string]any{
			"status": atcl.Status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("sync status failed, perhaps article[id: %d] doesn't belongs to this user[id: %d]", atcl.Id, atcl.AuthorId)
		}

		res = tx.Model(&PublishedArticle{}).Where("id = ?", atcl.Id).Updates(map[string]any{
			"status": atcl.Status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("sync status failed, perhaps article[id: %d] doesn't belongs to this user[id: %d]", atcl.Id, atcl.AuthorId)
		}

		return nil
	})
}
