//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/mercury/oauth2/grpc"
	"github.com/tsukaychan/mercury/oauth2/ioc"
	"github.com/tsukaychan/mercury/pkg/wego"
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
