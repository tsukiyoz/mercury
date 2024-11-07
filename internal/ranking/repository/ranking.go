package repository

import (
	"context"

	cache2 "github.com/lazywoo/mercury/internal/ranking/repository/cache"

	"github.com/lazywoo/mercury/internal/article/domain"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, atcls []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

var _ RankingRepository = (*RankingCachedRepository)(nil)

type RankingCachedRepository struct {
	redis *cache2.RankingRedisCache
	local *cache2.RankingLocalCache
}

func NewRankingCachedRepository(redisCache *cache2.RankingRedisCache, localCache *cache2.RankingLocalCache) RankingRepository {
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
