package ioc

import (
	articlev1 "github.com/lazywoo/mercury/pkg/api/article/v1"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitArticleRpcClient(etcdCli *clientv3.Client) articlev1.ArticleServiceClient {
	type config struct {
		Target string `yaml:"target"`
		Secure bool   `yaml:"secure"`
	}
	var cfg config
	err := viper.UnmarshalKey("grpc.client.article", &cfg)
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
	client := articlev1.NewArticleServiceClient(conn)
	return client
}
