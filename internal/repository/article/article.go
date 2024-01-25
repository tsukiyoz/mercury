package article

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
}

type CachedArticleRepository struct {
	articleDao articleDao.ArticleDAO

	// V1
	authorDao articleDao.AuthorDAO
	readerDao articleDao.ReaderDAO
}

func NewCachedArticleRepository(articleDao articleDao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		articleDao: articleDao,
	}
}

func (r *CachedArticleRepository) domainToEntity(article domain.Article) articleDao.Article {
	return articleDao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	}
}

func (r *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	return r.articleDao.UpdateById(ctx, articleDao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	})
}

func (r *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return r.articleDao.Insert(ctx, articleDao.Article{
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	})
}

func (r *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	return r.articleDao.Sync(ctx, r.domainToEntity(article))
}

func (r *CachedArticleRepository) SyncV1(ctx context.Context, article domain.Article) (int64, error) {
	var (
		id            = article.Id
		err           error
		articleEntity = r.domainToEntity(article)
	)
	if id > 0 {
		err = r.authorDao.UpdateById(ctx, articleEntity)
	} else {
		id, err = r.authorDao.Insert(ctx, articleEntity)
	}
	if err != nil {
		return id, err
	}

	err = r.readerDao.Upsert(ctx, articleEntity)
	return id, err
}
