package cache

import (
	"context"

	"github.com/tsukaychan/mercury/article/domain"
)

type RankingCache interface {
	Set(ctx context.Context, atcl []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}
