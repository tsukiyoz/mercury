package cache

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/lazywoo/mercury/interactive/domain"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/incr_cnt.lua
var luaIncrCnt string

//go:embed lua/batch_incr_cnt.lua
var luaBatchIncrCnt string

const (
	fieldReadCnt     = "read_cnt"
	fieldLikeCnt     = "like_cnt"
	fieldFavoriteCnt = "favorite_cnt"
)

//go:generate mockgen -source=./interactive.go -package=cachemocks -destination=mocks/interactive.mock.go InteractiveCache
type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCntIfPresent(ctx context.Context, biz string, bizIds []int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrFavoriteCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrFavoriteCntIfPresent(ctx context.Context, biz string, bizId int64) error
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

func (cache *RedisInteractiveCache) BatchIncrReadCntIfPresent(ctx context.Context, biz string, bizIds []int64) error {
	var keys []string
	for _, id := range bizIds {
		keys = append(keys, cache.key(biz, id))
	}
	return cache.client.Eval(ctx, luaBatchIncrCnt, keys, fieldReadCnt, 1).Err()
}

func (cache *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldLikeCnt, 1).Err()
}

func (cache *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldLikeCnt, -1).Err()
}

func (cache *RedisInteractiveCache) IncrFavoriteCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldFavoriteCnt, 1).Err()
}

func (cache *RedisInteractiveCache) DecrFavoriteCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldFavoriteCnt, -1).Err()
}

func (cache *RedisInteractiveCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	cnts, err := cache.client.HMGet(ctx, cache.key(biz, bizId), fieldReadCnt, fieldLikeCnt, fieldFavoriteCnt).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	var intr domain.Interactive
	if cnts[0] == nil || cnts[1] == nil || cnts[2] == nil {
		return domain.Interactive{}, ErrKeyNotExist
	}

	intr.ReadCnt, _ = strconv.ParseInt(cnts[0].(string), 10, 64)
	intr.LikeCnt, _ = strconv.ParseInt(cnts[1].(string), 10, 64)
	intr.FavoriteCnt, _ = strconv.ParseInt(cnts[2].(string), 10, 64)

	return intr, nil
}

func (cache *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	key := cache.key(biz, bizId)
	err := cache.client.HMSet(ctx, key,
		fieldReadCnt, intr.ReadCnt,
		fieldLikeCnt, intr.LikeCnt,
		fieldFavoriteCnt, intr.FavoriteCnt,
	).Err()
	if err != nil {
		return err
	}

	return cache.client.Expire(ctx, key, time.Minute*15).Err()
}
