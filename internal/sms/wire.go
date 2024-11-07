//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/internal/sms/grpc"
	"github.com/lazywoo/mercury/internal/sms/ioc"
	"github.com/lazywoo/mercury/pkg/wego"
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
