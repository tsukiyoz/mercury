package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/tsukaychan/webook/internal/domain"
	"time"
)

var ErrKeyNotExist = redis.Nil

var _ UserCache = (*UserRedisCache)(nil)

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
	Delete(ctx context.Context, id int64) error
}

type UserRedisCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func (cache *UserRedisCache) Get(ctx context.Context, id int64) (domain.User, error) {
	val, err := cache.client.Get(ctx, cache.key(id)).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *UserRedisCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.key(u.Id), val, cache.expiration).Err()
}

func (cache *UserRedisCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func (cache *UserRedisCache) Delete(ctx context.Context, id int64) error {
	return cache.client.Del(ctx, cache.key(id)).Err()
}

func NewUserRedisCache(client redis.Cmdable) UserCache {
	return &UserRedisCache{
		client:     client,
		expiration: time.Second * 3,
	}
}
