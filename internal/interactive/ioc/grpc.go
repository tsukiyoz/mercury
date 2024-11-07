package ioc

import (
	igrpc "github.com/lazywoo/mercury/internal/interactive/grpc"
	"github.com/lazywoo/mercury/pkg/grpcx"
	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGRPCxServer(intrSrv *igrpc.InteractiveServiceServer, l logger.Logger) *grpcx.Server {
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

	server := grpc.NewServer()
	intrSrv.Register(server)

	return grpcx.NewServer(server, "interactive", cfg.Port, []string{cfg.Etcd}, cfg.TTL, l)
}
