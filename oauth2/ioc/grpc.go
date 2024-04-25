package ioc

import (
	"github.com/spf13/viper"
	igrpc "github.com/tsukaychan/mercury/oauth2/grpc"
	"github.com/tsukaychan/mercury/pkg/grpcx"
	"github.com/tsukaychan/mercury/pkg/logger"
	"google.golang.org/grpc"
)

func InitGRPCxServer(server *igrpc.OAuth2ServiceServer, l logger.Logger) *grpcx.Server {
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
	server.Register(srv)
	return grpcx.NewServer(srv, "oauth2", cfg.Port, []string{cfg.Etcd}, cfg.TTL, l)
}
