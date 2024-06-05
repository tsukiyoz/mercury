package interceptor

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Builder struct{}

func (bdr *Builder) PeerName(ctx context.Context) string {
	return bdr.grpcHeaderValue(ctx, "app")
}

func (bdr *Builder) PeerIP(ctx context.Context) string {
	clientIP := bdr.grpcHeaderValue(ctx, "client-ip")
	if clientIP != "" {
		return clientIP
	}
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	if pr.Addr == net.Addr(nil) {
		return ""
	}
	addrs := strings.Split(pr.Addr.String(), ":")
	if len(addrs) > 1 {
		return addrs[0]
	}
	return ""
}

func (bdr *Builder) grpcHeaderValue(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return strings.Join(md.Get(key), ";")
}
