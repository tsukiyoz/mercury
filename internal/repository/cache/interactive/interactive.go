package cache

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tsukaychan/webook/internal/domain"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt     string
	ErrKeyNotExist = redis.Nil
)

const (
	fieldReadCnt    = "read_cnt"
	fieldLikeCnt    = "like_cnt"
	fieldCollectCnt = "collect_cnt"
)

//go:generate mockgen -source=./interactive.go -package=cachemocks -destination=mocks/interactive.mock.go InteractiveCache
type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
}

var _ InteractiveCache = (*RedisInteractiveCache)(nil)

type RedisInteractiveCache struct {
	client redis.Cmdable
}

func NewRedisInteractiveCache(client redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{
		client: client,
	}
}

func (cache *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}

func (cache *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldReadCnt, 1).Err()
}

func (cache *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldLikeCnt, 1).Err()
}

func (cache *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldLikeCnt, -1).Err()
}

func (cache *RedisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldCollectCnt, 1).Err()
}

//func (cache *RedisInteractiveCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
//	data, err := cache.client.HGetAll(ctx, cache.key(biz, bizId)).Result()
//	if err != nil {
//		return domain.Interactive{}, err
//	}
//
//	if len(data) == 0 {
//		return domain.Interactive{}, ErrKeyNotExist
//	}
//
//	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
//	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
//	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)
//
//	return domain.Interactive{
//		CollectCnt: collectCnt,
//		LikeCnt:    likeCnt,
//		ReadCnt:    readCnt,
//	}, nil
//}

func (cache *RedisInteractiveCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	cnts, err := cache.client.HMGet(ctx, cache.key(biz, bizId), fieldReadCnt, fieldLikeCnt, fieldCollectCnt).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	var intr domain.Interactive
	if cnts[0] == nil || cnts[1] == nil || cnts[2] == nil {
		return domain.Interactive{}, ErrKeyNotExist
	}

	intr.ReadCnt, _ = cnts[0].(int64)
	intr.LikeCnt, _ = cnts[1].(int64)
	intr.LikeCnt, _ = cnts[2].(int64)

	return intr, nil
}

func (cache *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	key := cache.key(biz, bizId)
	err := cache.client.HMSet(ctx, key,
		fieldReadCnt, intr.ReadCnt,
		fieldLikeCnt, intr.LikeCnt,
		fieldCollectCnt, intr.CollectCnt,
	).Err()
	if err != nil {
		return err
	}

	return cache.client.Expire(ctx, key, time.Minute*15).Err()
}
