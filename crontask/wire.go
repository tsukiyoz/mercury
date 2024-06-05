//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/crontask/grpc"
	"github.com/lazywoo/mercury/crontask/ioc"
	"github.com/lazywoo/mercury/crontask/repository"
	"github.com/lazywoo/mercury/crontask/repository/dao"
	"github.com/lazywoo/mercury/crontask/service"
	"github.com/lazywoo/mercury/pkg/wego"
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

func InitAPP() *wego.App {
	wire.Build(
		thirdProviderSet,
		ioc.InitGRPCxServer,
		svcProviderSet,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
