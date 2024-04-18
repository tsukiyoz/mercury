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
	ioc.InitSrcDB,
	ioc.InitDstDB,
	ioc.InitDualWritePool,
	ioc.InitDualWriteDB,
	//ioc.initDB,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitLogger,
	ioc.NewSyncProducer,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

var migratorSet = wire.NewSet(
	ioc.InitMigratorProducer,
	ioc.InitFixDataConsumer,
	ioc.InitMigratorWeb,
)

func InitApp() *App {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		migratorSet,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
