package user

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

var ErrKeyNotExist = redis.Nil

var _ UserCache = (*UserRedisCache)(nil)

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
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

func NewUserRedisCache(client redis.Cmdable) UserCache {
	return &UserRedisCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}
