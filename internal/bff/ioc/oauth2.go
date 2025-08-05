package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	oauth2v1 "github.com/tsukiyo/mercury/api/gen/oauth2/v1"
	"github.com/tsukiyo/mercury/internal/bff/web"
)

func InitOAuth2Client(etcdCli *clientv3.Client) oauth2v1.Oauth2ServiceClient {
	type Config struct {
		Target string `json:"target"`
		Secure bool   `json:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.oauth2", &cfg)
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
	return oauth2v1.NewOauth2ServiceClient(cc)
}

func InitWechatHandlerConfig() web.WechatHandlerConfig {
	type Config struct {
		Secure   bool `yaml:"secure"`
		HTTPOnly bool `yaml:"http_only"`
	}
	var cfg Config
	err := viper.UnmarshalKey("http", &cfg)
	if err != nil {
		panic(err)
	}
	return web.WechatHandlerConfig{
		Secure:   cfg.Secure,
		HTTPOnly: cfg.HTTPOnly,
	}
}
