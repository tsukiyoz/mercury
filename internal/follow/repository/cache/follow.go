package cache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/lazywoo/mercury/internal/follow/domain"
)

var ErrKeyNotExist = redis.Nil

var _ FollowCache = (*RedisFollowCache)(nil)

type RedisFollowCache struct {
	client redis.Cmdable
}

func NewRedisFollowCache(client redis.Cmdable) FollowCache {
	return &RedisFollowCache{
		client: client,
	}
}

const (
	fieldFollowerCnt = "follower_cnt"
	fieldFolloweeCnt = "followee_cnt"
)

func (cache *RedisFollowCache) staticsKey(uid int64) string {
	return fmt.Sprintf("follow:statics:%d", uid)
}

func (cache *RedisFollowCache) GetStatics(ctx context.Context, uid int64) (domain.Statics, error) {
	data, err := cache.client.HGetAll(ctx, cache.staticsKey(uid)).Result()
	if err != nil {
		return domain.Statics{}, err
	}
	if len(data) == 0 {
		return domain.Statics{}, ErrKeyNotExist
	}
	followerCnt, _ := strconv.ParseInt(data[fieldFollowerCnt], 10, 64)
	followeeCnt, _ := strconv.ParseInt(data[fieldFolloweeCnt], 10, 64)

	return domain.Statics{
		FolloweeCount: followeeCnt,
		FollowerCount: followerCnt,
	}, nil
}

func (cache *RedisFollowCache) SetStatics(ctx context.Context, uid int64, s domain.Statics) error {
	key := cache.staticsKey(uid)
	return cache.client.HMSet(ctx, key, fieldFolloweeCnt, s.FolloweeCount, fieldFollowerCnt, s.FollowerCount).Err()
}

func (cache *RedisFollowCache) updateStatics(ctx context.Context, r domain.Relation, delta int64) error {
	tx := cache.client.TxPipeline()
	tx.HIncrBy(ctx, cache.staticsKey(r.Followee), fieldFolloweeCnt, delta)
	tx.HIncrBy(ctx, cache.staticsKey(r.Follower), fieldFollowerCnt, delta)
	_, err := tx.Exec(ctx)
	return err
}

func (cache *RedisFollowCache) Follow(ctx context.Context, r domain.Relation) error {
	return cache.updateStatics(ctx, r, 1)
}

func (cache *RedisFollowCache) CancelFollow(ctx context.Context, r domain.Relation) error {
	return cache.updateStatics(ctx, r, -1)
}
