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

	Upsert(ctx context.Context, article PublishArticle) error
}

// Article Production Library
type Article struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Title    string `gorm:"type=varchar(1024)"`
	Content  string `gorm:"blob"`
	AuthorId int64  `gorm:"index=aid_ctime"`
	Ctime    int64  `gorm:"index=aid_ctime"`
	Utime    int64
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	err := dao.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.Utime = now
	res := dao.db.WithContext(ctx).Model(&article).
		Where("id=? AND author_id=?", article.Id, article.AuthorId).
		Updates(map[string]any{
			"title":   article.Title,
			"content": article.Content,
			"utime":   article.Utime,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("update article failed, perhaps invalid id: id %d author_id %d", article.Id, article.AuthorId)
	}
	return nil
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, article Article) (int64, error) {
	var id = article.Id
	err := dao.db.Transaction(func(tx *gorm.DB) (err error) {
		txDao := NewGORMArticleDAO(tx)

		if article.Id > 0 {
			err = txDao.UpdateById(ctx, article)
		} else {
			id, err = txDao.Insert(ctx, article)
		}
		if err != nil {
			return err
		}

		article.Id = id
		return txDao.Upsert(ctx, PublishArticle{article})
	})
	return id, err
}

func (dao *GORMArticleDAO) Upsert(ctx context.Context, article PublishArticle) error {
	now := time.Now().UnixMilli()
	article.Utime, article.Ctime = now, now
	err := dao.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   article.Title,
			"content": article.Content,
			"utime":   article.Utime,
		}),
	}).Create(&article).Error
	return err
}
