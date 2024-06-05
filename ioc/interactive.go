package ioc

import (
	"github.com/fsnotify/fsnotify"
	interactivev1 "github.com/lazywoo/mercury/api/proto/gen/interactive/v1"
	"github.com/lazywoo/mercury/interactive/service"
	"github.com/lazywoo/mercury/internal/web/client"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitInteractiveClient return grpc client
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
	cc, err := grpc.Dial(cfg.Target, opts...)
	if err != nil {
		panic(err)
	}
	return interactivev1.NewInteractiveServiceClient(cc)
}

// InitTrafficControlInteractiveGRPCClient return grpc client with traffic control
func InitTrafficControlInteractiveGRPCClient(svc service.InteractiveService) interactivev1.InteractiveServiceClient {
	type Config struct {
		Target    string
		Secure    bool
		Threshold int32
	}

	// remote
	var cfg Config
	viper.UnmarshalKey("grpc.client.interactive", &cfg)
	var opts []grpc.DialOption
	if cfg.Secure {
		// TODO HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Target, opts...)
	if err != nil {
		panic(err)
	}
	remote := interactivev1.NewInteractiveServiceClient(cc)

	// local
	local := client.NewInteractiveLocalAdapter(svc)

	intrCli := client.NewInteractiveClient(remote, local, 100)
	viper.OnConfigChange(func(in fsnotify.Event) {
		var cfg Config
		err = viper.UnmarshalKey("grpc.client.interactive", &cfg)
		if err != nil {
			// log
		}
		intrCli.UpdateThreshold(cfg.Threshold)
	})
	return intrCli
}
