package ratelimit

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tsukaychan/mercury/pkg/ratelimit"

	"google.golang.org/grpc"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
}

func (i *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := i.limiter.Limit(ctx, i.key)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "internal error")
		}
		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limited")
		}
		return handler(ctx, req)
	}
}

func (i *InterceptorBuilder) BuildClientInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		return handler(ctx, req)
	}
}
