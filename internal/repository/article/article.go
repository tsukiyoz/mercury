package article

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, atcl domain.Article) (int64, error)
	Update(ctx context.Context, atcl domain.Article) error
	Sync(ctx context.Context, atcl domain.Article) (int64, error)
	SyncStatus(ctx context.Context, atcl domain.Article) error
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

func (r *CachedArticleRepository) Update(ctx context.Context, atcl domain.Article) error {
	return r.articleDao.UpdateById(ctx, r.domainToEntity(atcl))
}

func (r *CachedArticleRepository) Create(ctx context.Context, atcl domain.Article) (int64, error) {
	return r.articleDao.Insert(ctx, r.domainToEntity(atcl))
}

func (r *CachedArticleRepository) Sync(ctx context.Context, atcl domain.Article) (int64, error) {
	return r.articleDao.Sync(ctx, r.domainToEntity(atcl))
}

func (r *CachedArticleRepository) SyncStatus(ctx context.Context, atcl domain.Article) error {
	return r.articleDao.SyncStatus(ctx, r.domainToEntity(atcl))
}
