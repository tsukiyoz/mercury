package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"

	gatewayv1 "github.com/tsukiyo/mercury/api/gen/gateway/v1"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

// command-line options:
// gRPC server endpoint
var grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:9090", "gRPC server endpoint")

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = gatewayv1.RegisterGatewayServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(":8081", mux)
}

type GatewayGRPCServer struct {
	gatewayv1.UnimplementedGatewayServiceServer
}

func (srv *GatewayGRPCServer) Echo(ctx context.Context, req *gatewayv1.EchoRequest) (*gatewayv1.EchoResponse, error) {
	fmt.Println("Got req value: ", req.Value)
	return &gatewayv1.EchoResponse{
		Value: req.Value,
	}, nil
}

func (s *GatewayGRPCServer) Check(ctx context.Context, in *healthv1.HealthCheckRequest) (*healthv1.HealthCheckResponse, error) {
	return &healthv1.HealthCheckResponse{Status: healthv1.HealthCheckResponse_SERVING}, nil
}

func (s *GatewayGRPCServer) Watch(in *healthv1.HealthCheckRequest, _ healthv1.Health_WatchServer) error {
	// Example of how to register both methods but only implement the Check method.
	return status.Error(codes.Unimplemented, "unimplemented")
}

func NewGatewayGRPCServer() *GatewayGRPCServer {
	return &GatewayGRPCServer{}
}

func main() {
	flag.Parse()

	go func() {
		srv := grpc.NewServer()
		gatewaysrv := NewGatewayGRPCServer()
		gatewayv1.RegisterGatewayServiceServer(srv, gatewaysrv)
		healthv1.RegisterHealthServer(srv, gatewaysrv)
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			panic(err)
		}
		srv.Serve(lis)
	}()

	if err := run(); err != nil {
		grpclog.Fatal(err)
	}
}
