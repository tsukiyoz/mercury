package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tsukiyo/mercury/internal/article/domain"

	"github.com/redis/go-redis/v9"
)

var _ RankingCache = (*RankingRedisCache)(nil)

type RankingRedisCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func NewRankingRedisCache(client redis.Cmdable) *RankingRedisCache {
	return &RankingRedisCache{client: client, key: "ranking:article", expiration: time.Minute * 3}
}

func (cache *RankingRedisCache) Set(ctx context.Context, atcls []domain.Article) error {
	for _, atcl := range atcls {
		atcl.Content = atcl.Abstract()
	}
	bs, err := json.Marshal(atcls)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.key, bs, cache.expiration).Err()
}

func (cache *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	bs, err := cache.client.Get(ctx, cache.key).Bytes()
	if err != nil {
		return nil, err
	}

	var atcls []domain.Article
	err = json.Unmarshal(bs, &atcls)
	if err != nil {
		return nil, err
	}

	return atcls, nil
}
