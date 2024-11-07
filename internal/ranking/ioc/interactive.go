package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	interactivev1 "github.com/lazywoo/mercury/pkg/api/interactive/v1"
)

func InitInteractiveRpcClient(etcdCli *clientv3.Client) interactivev1.InteractiveServiceClient {
	type config struct {
		Target string `yaml:"target"`
		Secure bool   `yaml:"secure"`
	}
	var cfg config
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
	conn, err := grpc.NewClient(cfg.Target, opts...)
	if err != nil {
		panic(err)
	}
	client := interactivev1.NewInteractiveServiceClient(conn)
	return client
}
