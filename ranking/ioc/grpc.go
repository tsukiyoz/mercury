package ioc

import (
	"github.com/spf13/viper"
	"github.com/tsukaychan/mercury/pkg/grpcx"
	"github.com/tsukaychan/mercury/pkg/logger"
	igrpc "github.com/tsukaychan/mercury/ranking/grpc"
	"google.golang.org/grpc"
)

func InitGRPCxServer(ranking *igrpc.RankingServiceServer, l logger.Logger) *grpcx.Server {
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
	ranking.Register(srv)
	return grpcx.NewServer(srv, "ranking", cfg.Port, []string{cfg.Etcd}, cfg.TTL, l)
}
