package ioc

import (
	"github.com/spf13/viper"
	grpc2 "github.com/tsukaychan/mercury/interactive/grpc"
	"github.com/tsukaychan/mercury/pkg/grpcx"
	"github.com/tsukaychan/mercury/pkg/logger"
	"google.golang.org/grpc"
)

func InitGRPCxServer(intrSrv *grpc2.InteractiveServiceServer, l logger.Logger) *grpcx.Server {
	type Config struct {
		Port int    `yaml:"port"`
		Etcd string `yaml:"etcd"`
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	intrSrv.Register(server)

	return grpcx.NewServer(server, "interactive", cfg.Port, []string{cfg.Etcd}, l)
}
