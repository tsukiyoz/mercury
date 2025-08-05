//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/tsukiyo/mercury/internal/sms/grpc"
	"github.com/tsukiyo/mercury/internal/sms/ioc"
	"github.com/tsukiyo/mercury/pkg/app"
)

var thirdProviderSet = wire.NewSet(
	// ioc.InitLogger,
	ioc.InitFileLogger,
)

func InitAPP() *app.App {
	wire.Build(
		thirdProviderSet,
		ioc.InitService,
		grpc.NewSmsServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(app.App), "GRPCServer"),
	)
	return new(app.App)
}
