package repository

import (
	"context"
	"time"

	"github.com/tsukiyo/mercury/internal/article/domain"
	"github.com/tsukiyo/mercury/internal/article/repository/cache"
	"github.com/tsukiyo/mercury/internal/article/repository/dao"

	"github.com/ecodeclub/ekit/slice"

	"github.com/tsukiyo/mercury/pkg/logger"
)

//go:generate mockgen -source=./article.go -package=repomocks -destination=mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, atcl domain.Article) (int64, error)
	Update(ctx context.Context, atcl domain.Article) error
	List(ctx context.Context, authorId int64, offset int, limit int) ([]domain.Article, error)
	ListPub(ctx context.Context, utime time.Time, offset int, limit int) ([]domain.Article, error)
	Sync(ctx context.Context, atcl domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id, authorId int64, status domain.ArticleStatus) error
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

var _ ArticleRepository = (*CachedArticleRepository)(nil)

type CachedArticleRepository struct {
	articleDAO   dao.ArticleDAO
	articleCache cache.ArticleCache
	logger       logger.Logger
}

func NewCachedArticleRepository(articleDao dao.ArticleDAO, articleCache cache.ArticleCache, l logger.Logger) ArticleRepository {
	return &CachedArticleRepository{
		articleDAO:   articleDao,
		articleCache: articleCache,
		logger:       l,
	}
}

func (repo *CachedArticleRepository) domainToEntity(atcl domain.Article) dao.Article {
	return dao.Article{
		Id:       atcl.Id,
		Title:    atcl.Title,
		Content:  atcl.Content,
		AuthorId: atcl.Author.Id,
		Status:   atcl.Status.ToUint8(),
		Ctime:    atcl.Ctime.UnixMilli(),
		Utime:    atcl.Utime.UnixMilli(),
	}
}

func (repo *CachedArticleRepository) entityToDomain(atcl dao.Article) domain.Article {
	return domain.Article{
		Id:      atcl.Id,
		Title:   atcl.Title,
		Content: atcl.Content,
		Author: domain.Author{
			Id: atcl.AuthorId,
		},
		Status: domain.ArticleStatus(atcl.Status),
		Ctime:  time.UnixMilli(atcl.Ctime),
		Utime:  time.UnixMilli(atcl.Utime),
	}
}

func (repo *CachedArticleRepository) Create(ctx context.Context, atcl domain.Article) (int64, error) {
	id, err := repo.articleDAO.Insert(ctx, repo.domainToEntity(atcl))
	if err != nil {
		return 0, err
	}
	err = repo.articleCache.DelFirstPage(ctx, atcl.Author.Id)
	if err != nil {
		repo.logger.Error("delete first page redis failed", logger.Int64("author_id", atcl.Author.Id), logger.Error(err))
	}
	return id, nil
}

func (repo *CachedArticleRepository) Update(ctx context.Context, atcl domain.Article) error {
	err := repo.articleDAO.UpdateById(ctx, repo.domainToEntity(atcl))
	if err != nil {
		return err
	}

	err = repo.articleCache.DelFirstPage(ctx, atcl.Author.Id)
	if err != nil {
		repo.logger.Error("delete first page redis failed", logger.Int64("author_id", atcl.Author.Id), logger.Error(err))
	}
	return err
}

func (repo *CachedArticleRepository) List(ctx context.Context, authorId int64, offset int, limit int) ([]domain.Article, error) {
	if offset == 0 && limit <= 100 {
		data, err := repo.articleCache.GetFirstPage(ctx, authorId)
		if err == nil {
			go func() {
				err := repo.preCache(ctx, data)
				if err != nil {
					repo.logger.Error("pre redis articles failed", logger.Int64("author_id", authorId), logger.Error(err))
				}
			}()
			return data[:min(len(data), limit)], err
		}
	}
	atcls, err := repo.articleDAO.GetByAuthor(ctx, authorId, offset, limit)
	if err != nil {
		return nil, err
	}
	data, err := slice.Map[dao.Article, domain.Article](atcls, func(idx int, src dao.Article) domain.Article {
		return repo.entityToDomain(src)
	}), nil

	go func() {
		err = repo.preCache(ctx, data)
	}()

	err = repo.articleCache.SetFirstPage(ctx, authorId, data)
	if err != nil {
		repo.logger.Error("write back redis failure", logger.Int64("author_id", authorId), logger.Error(err))
	}
	return data, nil
}

func (repo *CachedArticleRepository) ListPub(ctx context.Context, utime time.Time, offset int, limit int) ([]domain.Article, error) {
	pubAtcls, err := repo.articleDAO.ListPubByUtime(ctx, utime, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map[dao.PublishedArticle, domain.Article](pubAtcls, func(idx int, src dao.PublishedArticle) domain.Article {
		return repo.entityToDomain(dao.Article(src))
	}), nil
}

func (repo *CachedArticleRepository) Sync(ctx context.Context, atcl domain.Article) (int64, error) {
	id, err := repo.articleDAO.Sync(ctx, repo.domainToEntity(atcl))
	if err != nil {
		return 0, err
	}
	go func() {
		if err := repo.articleCache.DelFirstPage(ctx, atcl.Author.Id); err != nil {
			repo.logger.Error("write page back to redis failure", logger.Int64("author_id", atcl.Author.Id), logger.Error(err))
		}
		if err := repo.articleCache.SetPub(ctx, atcl); err != nil {
			repo.logger.Error("write article back to redis failure", logger.Int64("id", atcl.Id), logger.Error(err))
		}
	}()
	return id, err
}

func (repo *CachedArticleRepository) SyncStatus(ctx context.Context, id, authorId int64, status domain.ArticleStatus) error {
	return repo.articleDAO.SyncStatus(ctx, id, authorId, status.ToUint8())
}

func (repo *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	cachedAtcl, err := repo.articleCache.Get(ctx, id)
	if err == nil {
		return cachedAtcl, err
	}
	res, err := repo.articleDAO.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.entityToDomain(res), nil
}

func (repo *CachedArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	if cachedAtcl, err := repo.articleCache.GetPub(ctx, id); err == nil {
		return cachedAtcl, nil
	}
	atcl, err := repo.articleDAO.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	//dbUser, err := repo.userRepo.FindById(ctx, atcl.AuthorId)
	//if err != nil {
	//	return domain.Article{}, err
	//}
	res := domain.Article{
		Id:      atcl.Id,
		Title:   atcl.Title,
		Content: atcl.Content,
		Status:  domain.ArticleStatus(atcl.Status),
		Author: domain.Author{
			Id: atcl.AuthorId,
			// Name: dbUser.NickName,
		},
		Ctime: time.UnixMilli(atcl.Ctime),
		Utime: time.UnixMilli(atcl.Utime),
	}
	go func() {
		if err = repo.articleCache.SetPub(ctx, res); err != nil {
			repo.logger.Error("redis published article failed", logger.Int64("article_id", res.Id), logger.Error(err))
		}
	}()
	return res, nil
}

// preCache will redis first article in page and returns error.
func (repo *CachedArticleRepository) preCache(ctx context.Context, atcls []domain.Article) error {
	const contentSizeThreshold = 1024 * 1024
	if len(atcls) > 0 && len(atcls[0].Content) <= contentSizeThreshold {
		if err := repo.articleCache.Set(ctx, atcls[0]); err != nil {
			repo.logger.Error("pre redis article failed", logger.Error(err))
		}
	}
	return nil
}
