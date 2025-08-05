//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/tsukiyo/mercury/internal/account/grpc"
	"github.com/tsukiyo/mercury/internal/account/ioc"
	"github.com/tsukiyo/mercury/internal/account/repository"
	"github.com/tsukiyo/mercury/internal/account/repository/dao"
	"github.com/tsukiyo/mercury/internal/account/service"
	"github.com/tsukiyo/mercury/pkg/app"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
)

var svcProviderSet = wire.NewSet(
	grpc.NewAccountServiceServer,
	service.NewAccountServiceServer,
	repository.NewAccountRepository,
	dao.NewAccountDAO,
)

func InitAPP() *app.App {
	wire.Build(
		ioc.InitGRPCxServer,
		svcProviderSet,
		thirdProviderSet,
		wire.Struct(new(app.App), "GRPCServer"),
	)
	return new(app.App)
}
