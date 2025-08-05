package cache

import (
	"context"

	"github.com/tsukiyo/mercury/internal/article/domain"
)

type RankingCache interface {
	Set(ctx context.Context, atcl []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}
