package repository

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
}

type CachedArticleRepository struct {
	articleDao dao.ArticleDAO
}

func NewCachedArticleRepository(articleDao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		articleDao: articleDao,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return c.articleDao.Insert(ctx, dao.Article{
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	})
}
