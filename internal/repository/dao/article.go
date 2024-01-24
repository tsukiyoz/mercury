package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
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

func (dao *GORMArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	err := dao.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}
