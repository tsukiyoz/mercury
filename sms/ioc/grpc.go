package ioc

import (
	"github.com/lazywoo/mercury/pkg/grpcx"
	"github.com/lazywoo/mercury/pkg/logger"
	igrpc "github.com/lazywoo/mercury/sms/grpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGRPCxServer(sms *igrpc.SmsServiceServer, l logger.Logger) *grpcx.Server {
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
	sms.Register(srv)
	return grpcx.NewServer(srv, "sms", cfg.Port, []string{cfg.Etcd}, cfg.TTL, l)
}
