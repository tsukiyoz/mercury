package otel

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/tsukaychan/mercury/pkg/grpcx/interceptor"

	"google.golang.org/grpc/status"

	"github.com/tsukaychan/mercury/pkg/logger"
	"google.golang.org/grpc"
)

type LoggingInterceptorBuilder struct {
	l logger.Logger
	interceptor.Builder
}

func (bdr *LoggingInterceptorBuilder) BuildLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		evt := "normal"
		start := time.Now()
		fields := make([]logger.Field, 0, 20)
		defer func() {
			cost := time.Since(start)
			if rec := recover(); rec != nil {
				switch typ := rec.(type) {
				case error:
					err = typ
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				evt = "recover"
				err = status.New(codes.Internal, "panic, err: "+err.Error()).Err()
			}
			st, _ := status.FromError(err)
			fields = append(fields,
				logger.Int64("cost", cost.Milliseconds()),
				logger.String("type", "unary"),
				logger.String("method", info.FullMethod),
				logger.String("code", st.Code().String()),
				logger.String("code_msg", st.Message()),
				logger.String("peer", bdr.PeerName(ctx)),
				logger.String("peer_ip", bdr.PeerIP(ctx)),
				logger.String("event", evt),
			)

			bdr.l.Info("rpc", fields...)
		}()
		resp, err = handler(ctx, req)
		return
	}
}
