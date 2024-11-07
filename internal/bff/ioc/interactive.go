package ioc

import (
	interactivev1 "github.com/lazywoo/mercury/pkg/api/interactive/v1"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitInteractiveClient(etcdCli *clientv3.Client) interactivev1.InteractiveServiceClient {
	type Config struct {
		Target string `json:"target"`
		Secure bool   `json:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.interactive", &cfg)
	if err != nil {
		panic(err)
	}
	rs, err := resolver.NewBuilder(etcdCli)
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{grpc.WithResolvers(rs)}
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient(cfg.Target, opts...)
	if err != nil {
		panic(err)
	}
	return interactivev1.NewInteractiveServiceClient(cc)
}
