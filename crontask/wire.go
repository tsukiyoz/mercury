//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/mercury/crontask/grpc"
	"github.com/tsukaychan/mercury/crontask/ioc"
	"github.com/tsukaychan/mercury/crontask/repository"
	"github.com/tsukaychan/mercury/crontask/repository/dao"
	"github.com/tsukaychan/mercury/crontask/service"
	"github.com/tsukaychan/mercury/pkg/wego"
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
