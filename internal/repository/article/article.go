package article

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
}

type CachedArticleRepository struct {
	articleDao dao.ArticleDAO
}

func NewCachedArticleRepository(articleDao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		articleDao: articleDao,
	}
}

func (r *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	return r.articleDao.UpdateById(ctx, dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	})
}

func (r *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return r.articleDao.Insert(ctx, dao.Article{
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	})
}
