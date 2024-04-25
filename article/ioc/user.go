package ioc

import (
	"github.com/spf13/viper"
	userv1 "github.com/tsukaychan/mercury/api/proto/gen/user/v1"
	"google.golang.org/grpc"
)

func InitUserRpcClient() userv1.UserServiceClient {
	type config struct {
		Addr string `yaml:"addr"`
	}
	var cfg config
	err := viper.UnmarshalKey("userGrpc", &cfg)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.Dial(cfg.Addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := userv1.NewUserServiceClient(conn)
	return client
}
