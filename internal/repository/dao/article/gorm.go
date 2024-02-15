package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

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
	res := dao.db.WithContext(ctx).Model(&atcl).
		Where("id = ? AND author_id = ?", atcl.Id, atcl.AuthorId).
		Updates(map[string]any{
			"title":   atcl.Title,
			"content": atcl.Content,
			"status":  atcl.Status,
			"utime":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("update article failed, perhaps invalid id: id %d author_id %d", atcl.Id, atcl.AuthorId)
	}
	return nil
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, authorId int64, offset, limit int) ([]Article, error) {
	var atcls []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", authorId).
		Offset(offset).Limit(limit).
		Order("utime DESC").
		Find(&atcls).Error
	return atcls, err
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var atcl Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", id).
		First(&atcl).Error
	return atcl, err
}

func (dao *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pubAtcl PublishedArticle
	err := dao.db.WithContext(ctx).Model(&PublishedArticle{}).
		Where("id = ?", id).
		First(&pubAtcl).Error
	return pubAtcl, err
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, atcl Article) (int64, error) {
	tx := dao.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer tx.Rollback()

	txDao := NewGORMArticleDAO(tx)
	var (
		id  = atcl.Id
		err error
	)
	if id == 0 {
		id, err = txDao.Insert(ctx, atcl)
	} else {
		err = txDao.UpdateById(ctx, atcl)
	}
	if err != nil {
		return 0, err
	}
	atcl.Id = id
	pubAtcl := PublishedArticle(atcl)
	pubAtcl.Utime, pubAtcl.Ctime = now, now
	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   atcl.Title,
			"content": atcl.Content,
			"status":  atcl.Status,
			"utime":   now,
		}),
	}).Create(&pubAtcl).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Error
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id, authorId int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, authorId).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&PublishedArticle{}).Where("id = ?", id).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		return nil
	})
}

func (dao *GORMArticleDAO) ListPubByUtime(ctx context.Context, utime time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var pubAtcls []PublishedArticle
	err := dao.db.WithContext(ctx).Model(&PublishedArticle{}).
		Order("utime DESC").Where("utime < ?", utime.UnixMilli()).
		Limit(limit).
		Offset(offset).
		Find(&pubAtcls).Error
	return pubAtcls, err
}
