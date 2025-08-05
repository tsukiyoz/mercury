package ratelimit

import (
	"context"
	"fmt"

	"github.com/tsukiyo/mercury/internal/sms/service"
	"github.com/tsukiyo/mercury/internal/sms/service/tencent"
	"github.com/tsukiyo/mercury/pkg/ratelimit"
)

var errLimited = fmt.Errorf("ratelimited")

type RateLimitSMSService struct {
	svc     service.Service
	limiter ratelimit.Limiter
}

func (s RateLimitSMSService) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	limited, err := s.limiter.Limit(ctx, tencent.LimitKey)
	if err != nil {
		return fmt.Errorf("sms ratelimiter limit err: %w", err)
	}
	if limited {
		return errLimited
	}
	err = s.svc.Send(ctx, tpl, target, args, values)
	return err
}

func NewRateLimitSMSService(svc service.Service, limiter ratelimit.Limiter) service.Service {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}
