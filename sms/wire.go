//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/pkg/wego"
	"github.com/lazywoo/mercury/sms/grpc"
	"github.com/lazywoo/mercury/sms/ioc"
)

var thirdProviderSet = wire.NewSet(
	// ioc.InitLogger,
	ioc.InitFileLogger,
)

func InitAPP() *wego.App {
	wire.Build(
		thirdProviderSet,
		ioc.InitService,
		grpc.NewSmsServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
