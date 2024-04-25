package ioc

import (
	"github.com/spf13/viper"
	igrpc "github.com/tsukaychan/mercury/comment/grpc"
	"github.com/tsukaychan/mercury/pkg/grpcx"
	"github.com/tsukaychan/mercury/pkg/logger"
	"google.golang.org/grpc"
)

func InitGRPCxServer(comment *igrpc.CommentServiceServer, l logger.Logger) *grpcx.Server {
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
	comment.Register(srv)
	return grpcx.NewServer(srv, "comment", cfg.Port, []string{cfg.Etcd}, cfg.TTL, l)
}
