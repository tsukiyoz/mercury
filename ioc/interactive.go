package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	interactivev1 "github.com/tsukaychan/mercury/api/proto/gen/interactive/v1"
	"github.com/tsukaychan/mercury/interactive/service"
	"github.com/tsukaychan/mercury/internal/web/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitInteractiveGRPCClient(svc service.InteractiveService) interactivev1.InteractiveServiceClient {
	type Config struct {
		Addr      string
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
	cc, err := grpc.Dial(cfg.Addr, opts...)
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
