package cache

import (
	"context"
	"errors"
	"time"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/tsukaychan/webook/internal/domain"
)

type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func NewRankingLocalCache() *RankingLocalCache {
	return &RankingLocalCache{
		topN:       atomicx.NewValue[[]domain.Article](),
		ddl:        atomicx.NewValueOf[time.Time](time.Now()),
		expiration: time.Minute * 3,
	}
}

func (cache *RankingLocalCache) Set(_ context.Context, atcls []domain.Article) error {
	cache.ddl.Store(time.Now().Add(time.Minute * 3))
	cache.topN.Store(atcls)
	return nil
}

func (cache *RankingLocalCache) Get(_ context.Context) ([]domain.Article, error) {
	atcls := cache.topN.Load()
	if len(atcls) == 0 || cache.ddl.Load().Before(time.Now()) {
		return nil, errors.New("local cache failure")
	}
	return atcls, nil
}

func (cache *RankingLocalCache) ForceGet(_ context.Context) ([]domain.Article, error) {
	return cache.topN.Load(), nil
}
