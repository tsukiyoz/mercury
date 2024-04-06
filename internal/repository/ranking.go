package repository

import (
	"context"

	cache "github.com/tsukaychan/mercury/internal/repository/cache/ranking"

	"github.com/tsukaychan/mercury/internal/domain"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, atcls []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

var _ RankingRepository = (*RankingCachedRepository)(nil)

type RankingCachedRepository struct {
	redis *cache.RankingRedisCache
	local *cache.RankingLocalCache
}

func NewRankingCachedRepository(redisCache *cache.RankingRedisCache, localCache *cache.RankingLocalCache) RankingRepository {
	return &RankingCachedRepository{
		redis: redisCache,
		local: localCache,
	}
}

func (repo *RankingCachedRepository) ReplaceTopN(ctx context.Context, atcls []domain.Article) error {
	_ = repo.local.Set(ctx, atcls)
	return repo.redis.Set(ctx, atcls)
}

func (repo *RankingCachedRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	atcls, err := repo.local.Get(ctx)
	if err == nil {
		return atcls, nil
	}
	atcls, err = repo.redis.Get(ctx)
	if err == nil {
		_ = repo.local.Set(ctx, atcls)
	} else {
		return repo.local.ForceGet(ctx)
	}
	return atcls, err
}
