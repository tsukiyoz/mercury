package circuitbreaker

import (
	"context"
	"errors"

	//"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
)

var ErrNotAllowed = errors.New("request failed due to circuit breaker triggered")

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if b.breaker.Allow() == nil {
			resp, err = handler(ctx, req)
			sts, ok := status.FromError(err)
			if !ok {
				b.breaker.MarkFailed()
				return resp, err
			}
			if sts != nil && sts.Code() == codes.Unavailable {
				b.breaker.MarkFailed()
			} else {
				b.breaker.MarkSuccess()
			}
			if err != nil {
				b.breaker.MarkFailed()
			} else {
				b.breaker.MarkSuccess()
			}
			return resp, err
		}
		b.breaker.MarkFailed()
		return nil, err
	}
}
