package otel

import (
	"context"

	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/go-kratos/kratos/v2/errors"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/metadata"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/tsukiyo/mercury/pkg/grpcx/interceptor"

	"google.golang.org/grpc"
)

type TraceInterceptorBuilder struct {
	interceptor.Builder

	serviceName string
	tracer      trace.Tracer
	propagator  propagation.TextMapPropagator
}

func NewTraceInterceptorBuilder(serviceName string, tracer trace.Tracer, propagator propagation.TextMapPropagator) *TraceInterceptorBuilder {
	return &TraceInterceptorBuilder{
		serviceName: serviceName,
		tracer:      tracer,
		propagator:  propagator,
	}
}

func (bdr *TraceInterceptorBuilder) BuildUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	propagator := bdr.propagator
	if propagator == nil {
		// global
		propagator = otel.GetTextMapPropagator()
	}
	tracer := bdr.tracer
	if tracer == nil {
		tracer = otel.GetTracerProvider().Tracer("github.com/tsukiyo/mercury/pkg/grpcx/interceptor/otel")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("client"),
	}
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		attrs = append(attrs, semconv.RPCMethodKey.String(method),
			semconv.NetPeerNameKey.String(bdr.serviceName),
		)
		ctx, span := tracer.Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)
		ctx = inject(ctx, propagator)
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
			span.End()
		}()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (bdr *TraceInterceptorBuilder) BuildUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	propagator := bdr.propagator
	if propagator == nil {
		propagator = otel.GetTextMapPropagator() // global
	}
	tracer := bdr.tracer
	if tracer == nil {
		tracer = otel.GetTracerProvider().Tracer("github.com/tsukiyo/mercury/pkg/grpcx/interceptor/otel")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("server"),
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = extract(ctx, propagator)
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attrs...),
		)

		span.SetAttributes(
			semconv.RPCMethodKey.String(info.FullMethod),
			semconv.NetPeerNameKey.String(bdr.PeerName(ctx)),
			attribute.Key("net.peer.ip").String(bdr.PeerIP(ctx)),
		)
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				} else {
					span.SetStatus(codes.Ok, "OK")
				}
			}
			span.End()
		}()
		return handler(ctx, req)
	}
}

func extract(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	return propagators.Extract(ctx, MetadataCarrier(md))
}

func inject(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	propagators.Inject(ctx, MetadataCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

type MetadataCarrier metadata.MD

func (mc MetadataCarrier) Get(key string) string {
	val := metadata.MD(mc).Get(key)
	if len(val) > 0 {
		return val[0]
	}
	return ""
}

func (mc MetadataCarrier) Set(key string, value string) {
	metadata.MD(mc).Set(key, value)
}

func (mc MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for key := range metadata.MD(mc) {
		keys = append(keys, key)
	}
	return keys
}
