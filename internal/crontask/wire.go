//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/tsukiyo/mercury/internal/crontask/grpc"
	"github.com/tsukiyo/mercury/internal/crontask/ioc"
	"github.com/tsukiyo/mercury/internal/crontask/repository"
	"github.com/tsukiyo/mercury/internal/crontask/repository/dao"
	"github.com/tsukiyo/mercury/internal/crontask/service"
	"github.com/tsukiyo/mercury/pkg/app"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
)

var svcProviderSet = wire.NewSet(
	grpc.NewCronJobServiceServer,
	service.NewTaskService,
	repository.NewPreemptTaskRepository,
	dao.NewGORMTaskDAO,
)

func InitAPP() *app.App {
	wire.Build(
		thirdProviderSet,
		ioc.InitGRPCxServer,
		svcProviderSet,
		wire.Struct(new(app.App), "GRPCServer"),
	)
	return new(app.App)
}
