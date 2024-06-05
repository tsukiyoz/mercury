package ratelimit

import "context"

//go:generate mockgen -source=./types.go -package=limitmocks -destination=mocks/ratelimit.mock.go Limiter
type Limiter interface {
	// Limit return is key limited
	Limit(ctx context.Context, key string) (bool, error)
}
