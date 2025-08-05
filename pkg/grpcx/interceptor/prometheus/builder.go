package prometheus

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/tsukiyo/mercury/pkg/grpcx/interceptor"
)

type InterceptorBuilder struct {
	interceptor.Builder

	Namespace string
	Subsystem string
}

func (b *InterceptorBuilder) BuildServer() grpc.UnaryServerInterceptor {
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: b.Namespace,
			Subsystem: b.Subsystem,
			Name:      "server_handle_seconds",
			Objectives: map[float64]float64{
				0.5:   0.01,
				0.9:   0.01,
				0.95:  0.01,
				0.99:  0.001,
				0.999: 0.0001,
			},
		},
		[]string{"type", "service", "method", "peer", "code"},
	)
	prometheus.Register(summary)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Milliseconds()
			srv, method := b.split(info.FullMethod)
			code := "OK"
			if err != nil {
				st, _ := status.FromError(err)
				code = st.Code().String()
			}
			summary.WithLabelValues("unary", srv, method, b.PeerName(ctx), code).Observe(float64(duration))
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func (b *InterceptorBuilder) split(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/")
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}
