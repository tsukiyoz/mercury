package service

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
}

var _ ArticleService = (*articleService)(nil)

type articleService struct {
	articleRepo repository.ArticleRepository
}

func NewArticleService(articleRepo repository.ArticleRepository) ArticleService {
	return &articleService{
		articleRepo: articleRepo,
	}
}

func (s *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	return s.articleRepo.Create(ctx, article)
}
