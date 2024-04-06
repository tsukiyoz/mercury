package ioc

import (
	"github.com/spf13/viper"
	grpc2 "github.com/tsukaychan/mercury/interactive/grpc"
	"github.com/tsukaychan/mercury/pkg/grpcx"
	"google.golang.org/grpc"
)

func InitGRPCxServer(intrSrv *grpc2.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `json:"addr"`
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil || cfg.Addr == "" {
		panic(err)
	}

	server := grpc.NewServer()
	intrSrv.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
