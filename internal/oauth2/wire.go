//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/internal/oauth2/grpc"
	"github.com/lazywoo/mercury/internal/oauth2/ioc"
	"github.com/lazywoo/mercury/pkg/wego"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitLogger,
)

func InitAPP() *wego.App {
	wire.Build(
		thirdProviderSet,
		ioc.InitWechatService,
		grpc.NewOAuth2ServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
