package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/slide_window.lua
var luaSlideWindow string

type RedisSlidingWindowLimiter struct {
	cmd redis.Cmdable
	// internal window size
	internal time.Duration
	// rate max requests number in the internal scope
	rate int
}

func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaSlideWindow, []string{key}, r.internal.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}

func (r *RedisSlidingWindowLimiter) Internal(internal time.Duration) {
	r.internal = internal
}

func (r *RedisSlidingWindowLimiter) Rate(rate int) {
	r.rate = rate
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, internal time.Duration, rate int) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		internal: internal,
		rate:     rate,
	}
}
