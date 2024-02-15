package dao

import (
	"context"
	"errors"
	"time"
)

var ErrPossibleIncorrectAuthor = errors.New("the user is attempting to manipulate non personal data")

type ArticleDAO interface {
	Insert(ctx context.Context, atcl Article) (int64, error)
	UpdateById(ctx context.Context, atcl Article) error
	GetByAuthor(ctx context.Context, authorId int64, offset, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	Sync(ctx context.Context, atcl Article) (int64, error)
	SyncStatus(ctx context.Context, id, authorId int64, status uint8) error
	ListPubByUtime(ctx context.Context, utime time.Time, offset int, limit int) ([]PublishedArticle, error)
}
