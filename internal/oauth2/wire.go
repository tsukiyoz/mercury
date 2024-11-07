//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/internal/oauth2/grpc"
	"github.com/lazywoo/mercury/internal/oauth2/ioc"
	"github.com/lazywoo/mercury/pkg/app"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitLogger,
)

func InitAPP() *app.App {
	wire.Build(
		thirdProviderSet,
		ioc.InitWechatService,
		grpc.NewOAuth2ServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(app.App), "GRPCServer"),
	)
	return new(app.App)
}
