//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/webook/interactive/events"
	"github.com/tsukaychan/webook/interactive/grpc"
	"github.com/tsukaychan/webook/interactive/ioc"
	"github.com/tsukaychan/webook/interactive/repository"
	"github.com/tsukaychan/webook/interactive/repository/cache"
	"github.com/tsukaychan/webook/interactive/repository/dao"
	"github.com/tsukaychan/webook/interactive/service"
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
