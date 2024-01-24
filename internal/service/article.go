package service

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository/article"
	"github.com/tsukaychan/webook/pkg/logger"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	PublishV1(ctx context.Context, article domain.Article) (int64, error)
}

var _ ArticleService = (*articleService)(nil)

type articleService struct {
	// V0
	articleRepo article.ArticleRepository

	// V1
	articleAuthorRepo article.ArticleAuthorRepository
	articleReaderRepo article.ArticleReaderRepository

	logger logger.Logger
}

func NewArticleService(articleRepo article.ArticleRepository) ArticleService {
	return &articleService{
		articleRepo: articleRepo,
	}
}

func NewArticleServiceV1(authorRepo article.ArticleAuthorRepository, readerRepo article.ArticleReaderRepository, logger logger.Logger) ArticleService {
	return &articleService{
		articleAuthorRepo: authorRepo,
		articleReaderRepo: readerRepo,
		logger:            logger,
	}
}

func (s *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	if article.Id > 0 {
		err := s.articleRepo.Update(ctx, article)
		return article.Id, err
	}

	return s.articleRepo.Create(ctx, article)
}

func (s *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *articleService) PublishV1(ctx context.Context, article domain.Article) (int64, error) {
	var (
		id  = article.Id
		err error
	)
	if article.Id > 0 {
		err = s.articleAuthorRepo.Update(ctx, article)
	} else {
		id, err = s.articleAuthorRepo.Create(ctx, article)
	}
	if err != nil {
		return 0, err
	}
	article.Id = id
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * time.Duration(i))
		id, err = s.articleReaderRepo.Save(ctx, article)
		if err == nil {
			break
		}
		s.logger.Error("partial failure, save into online library failed",
			logger.Int64("article_id", article.Id),
			logger.Error(err),
		)
	}
	if err != nil {
		s.logger.Error("total failure, save into online library failed",
			logger.Int64("article_id", article.Id),
			logger.Error(err),
		)
		// TODO Connect to the alarm system and handle it manually
		// use Canal for sync in the future
	}
	return id, err
}
