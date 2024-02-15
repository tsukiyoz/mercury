package repository

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
	"time"
)

//go:generate mockgen -source=./article.go -package=repomocks -destination=mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, atcl domain.Article) (int64, error)
	Update(ctx context.Context, atcl domain.Article) error
	List(ctx context.Context, author int64, offset int, limit int) ([]domain.Article, error)
	Sync(ctx context.Context, atcl domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id, authorId int64, status domain.ArticleStatus) error
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, utime time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	articleDao articleDao.ArticleDAO
}

func NewCachedArticleRepository(articleDao articleDao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		articleDao: articleDao,
	}
}

func (r *CachedArticleRepository) domainToEntity(atcl domain.Article) articleDao.Article {
	return articleDao.Article{
		Id:       atcl.Id,
		Title:    atcl.Title,
		Content:  atcl.Content,
		AuthorId: atcl.Author.Id,
		Status:   atcl.Status.ToUint8(),
	}
}

func (r *CachedArticleRepository) Create(ctx context.Context, atcl domain.Article) (int64, error) {
	return r.articleDao.Insert(ctx, r.domainToEntity(atcl))
}

func (r *CachedArticleRepository) Update(ctx context.Context, atcl domain.Article) error {
	return r.articleDao.UpdateById(ctx, r.domainToEntity(atcl))
}

func (r *CachedArticleRepository) List(ctx context.Context, author int64, offset int, limit int) ([]domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CachedArticleRepository) Sync(ctx context.Context, atcl domain.Article) (int64, error) {
	return r.articleDao.Sync(ctx, r.domainToEntity(atcl))
}

func (r *CachedArticleRepository) SyncStatus(ctx context.Context, id, authorId int64, status domain.ArticleStatus) error {
	return r.articleDao.SyncStatus(ctx, id, authorId, status.ToUint8())
}

func (r *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CachedArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CachedArticleRepository) ListPub(ctx context.Context, utime time.Time, offset int, limit int) ([]domain.Article, error) {
	//TODO implement me
	panic("implement me")
}
