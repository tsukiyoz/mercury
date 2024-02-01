package service

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository/article"
	"github.com/tsukaychan/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, atcl domain.Article) (int64, error)
	Publish(ctx context.Context, atcl domain.Article) (int64, error)
	Withdraw(ctx context.Context, atcl domain.Article) error
}

var _ ArticleService = (*articleService)(nil)

type articleService struct {
	articleRepo article.ArticleRepository
	logger      logger.Logger
}

func NewArticleService(articleRepo article.ArticleRepository, logger logger.Logger) ArticleService {
	return &articleService{
		articleRepo: articleRepo,
		logger:      logger,
	}
}

func (s *articleService) Save(ctx context.Context, atcl domain.Article) (int64, error) {
	atcl.Status = domain.ArticleStatusUnpublished
	if atcl.Id > 0 {
		err := s.articleRepo.Update(ctx, atcl)
		return atcl.Id, err
	}

	return s.articleRepo.Create(ctx, atcl)
}

func (s *articleService) Publish(ctx context.Context, atcl domain.Article) (int64, error) {
	atcl.Status = domain.ArticleStatusPublished
	return s.articleRepo.Sync(ctx, atcl)
}

func (s *articleService) Withdraw(ctx context.Context, atcl domain.Article) error {
	atcl.Status = domain.ArticleStatusPrivate
	return s.articleRepo.SyncStatus(ctx, atcl)
}
