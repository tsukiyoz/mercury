package lua

import (
	"context"
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/redis/go-redis/v9"
)

var (
	//go:embed batch_incr_cnt.lua
	luaBatchIncrCnt string
	client          redis.Cmdable
	biz             = "test_before"
	batchSize       = 5
)

func InitRedis() {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "for.nothing",
		DB:       1,
	})
}

func TestBatchIncrLua(t *testing.T) {
	InitRedis()

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	// defer cancel()
	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	assert.NoError(t, err)

	var keys []string
	for i := 0; i < batchSize; i++ {
		keys = append(keys, fmt.Sprintf("interactive:%s:%d", biz, i))
	}

	for _, key := range keys {
		err = client.HMSet(ctx, key,
			"read_cnt", 1,
			"like_cnt", 1,
			"favorite_cnt", 1,
		).Err()
		assert.NoError(t, err)

		err = client.Expire(ctx, key, time.Minute*3).Err()
		assert.NoError(t, err)
	}

	err = client.Eval(ctx, luaBatchIncrCnt, keys, "like_cnt", 1).Err()
	if err != nil {
		t.Logf("%v\n", err)
	}

	// results := client.HGetAll()
	for _, key := range keys {
		data, err := client.HMGet(ctx, key, "read_cnt", "like_cnt", "favorite_cnt").Result()
		assert.NoError(t, err)
		t.Logf("%v\n", data)
	}
}

func TestHMGetNonExists(t *testing.T) {
	InitRedis()

	ctx := context.Background()
	key := "non_exists"
	results, err := client.HMGet(ctx, key, "read_cnt", "like_cnt", "favorite_cnt").Result()
	assert.NoError(t, err)
	for _, result := range results {
		if result != nil {
			t.Logf("result: %v\n", result)
		}
	}
}
