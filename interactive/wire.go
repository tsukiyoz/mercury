//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/mercury/interactive/events"
	"github.com/tsukaychan/mercury/interactive/grpc"
	"github.com/tsukaychan/mercury/interactive/ioc"
	"github.com/tsukaychan/mercury/interactive/repository"
	"github.com/tsukaychan/mercury/interactive/repository/cache"
	"github.com/tsukaychan/mercury/interactive/repository/dao"
	"github.com/tsukaychan/mercury/interactive/service"
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitLogger,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitApp() *App {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
