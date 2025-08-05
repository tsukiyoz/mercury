package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	igrpc "github.com/tsukiyo/mercury/internal/user/grpc"
	"github.com/tsukiyo/mercury/pkg/grpcx"
	"github.com/tsukiyo/mercury/pkg/logger"
)

func InitGRPCxServer(user *igrpc.UserServiceServer, l logger.Logger) *grpcx.Server {
	type Config struct {
		Port int    `yaml:"port"`
		Etcd string `yaml:"etcd"`
		TTL  int64  `yaml:"ttl"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	user.Register(srv)
	return grpcx.NewServer(srv, "user", cfg.Port, []string{cfg.Etcd}, cfg.TTL, l)
}
