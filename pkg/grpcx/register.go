package grpcx

import "google.golang.org/grpc"

type Register interface {
	Register(srv *grpc.Server)
}
