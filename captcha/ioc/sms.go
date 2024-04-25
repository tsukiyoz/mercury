package ioc

import (
	"github.com/spf13/viper"
	smsv1 "github.com/tsukaychan/mercury/api/proto/gen/sms/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitSmsServiceClient() smsv1.SmsServiceClient {
	type config struct {
		Target string `yaml:"target"`
	}
	var cfg config
	err := viper.UnmarshalKey("grpc.client.sms", &cfg)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.NewClient(
		cfg.Target,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := smsv1.NewSmsServiceClient(conn)
	return client
}
