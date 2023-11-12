package ratelimit

import "context"

type Limiter interface {
	// Limit return is key limited
	Limit(ctx context.Context, key string) (bool, error)
}
