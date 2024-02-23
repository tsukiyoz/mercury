package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tsukaychan/webook/internal/domain"
)

//go:generate mockgen -source=./article.go -package=cachemocks -destination=mocks/article.mock.go ArticleCache
type ArticleCache interface {
	SetFirstPage(ctx context.Context, authorId int64, atcls []domain.Article) error
	DelFirstPage(ctx context.Context, authorId int64) error
	GetFirstPage(ctx context.Context, authorId int64) ([]domain.Article, error)

	Set(ctx context.Context, atcl domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)

	SetPub(ctx context.Context, atcl domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

func NewRedisArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{
		client: client,
	}
}

var _ ArticleCache = (*RedisArticleCache)(nil)

type RedisArticleCache struct {
	client redis.Cmdable
}

func (cache *RedisArticleCache) SetFirstPage(ctx context.Context, authorId int64, atcls []domain.Article) error {
	for i := range atcls {
		atcls[i].Content = atcls[i].Abstract()
	}
	bs, err := json.Marshal(atcls)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.firstPageKey(authorId), bs, time.Minute*10).Err()
}

func (cache *RedisArticleCache) firstPageKey(authorId int64) string {
	return fmt.Sprintf("article:first_page:%d", authorId)
}

func (cache *RedisArticleCache) DelFirstPage(ctx context.Context, authorId int64) error {
	return cache.client.Del(ctx, cache.firstPageKey(authorId)).Err()
}

func (cache *RedisArticleCache) GetFirstPage(ctx context.Context, authorId int64) ([]domain.Article, error) {
	bs, err := cache.client.Get(ctx, cache.firstPageKey(authorId)).Bytes()
	if err != nil {
		return nil, err
	}
	var atcls []domain.Article
	err = json.Unmarshal(bs, &atcls)
	return atcls, err
}

func (cache *RedisArticleCache) Set(ctx context.Context, atcl domain.Article) error {
	data, err := json.Marshal(atcl)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.authorArticleKey(atcl.Id), data, time.Second*10).Err()
}

func (cache *RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	var atcl domain.Article
	bs, err := cache.client.Get(ctx, cache.authorArticleKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	err = json.Unmarshal(bs, &atcl)
	return atcl, err
}

func (cache *RedisArticleCache) SetPub(ctx context.Context, atcl domain.Article) error {
	bs, err := json.Marshal(atcl)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.readerArticleKey(atcl.Id), bs, time.Minute*30).Err()
}

func (cache *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	var atcl domain.Article
	bs, err := cache.client.Get(ctx, cache.authorArticleKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	err = json.Unmarshal(bs, &atcl)
	return atcl, err
}

func (cache *RedisArticleCache) readerArticleKey(id int64) string {
	return fmt.Sprintf("article:reader:%d", id)
}

func (cache *RedisArticleCache) authorArticleKey(id int64) string {
	return fmt.Sprintf("article:author:%d", id)
}
