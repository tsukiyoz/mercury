package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	igrpc "github.com/lazywoo/mercury/internal/payment/grpc"
	"github.com/lazywoo/mercury/pkg/grpcx"
	"github.com/lazywoo/mercury/pkg/logger"
)

func InitGRPCxServer(payment *igrpc.WechatServiceServer, l logger.Logger) *grpcx.Server {
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
	payment.Register(srv)
	return grpcx.NewServer(srv, "payment", cfg.Port, []string{cfg.Etcd}, cfg.TTL, l)
}
