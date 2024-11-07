//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/internal/interactive/events"
	"github.com/lazywoo/mercury/internal/interactive/grpc"
	"github.com/lazywoo/mercury/internal/interactive/ioc"
	"github.com/lazywoo/mercury/internal/interactive/repository"
	"github.com/lazywoo/mercury/internal/interactive/repository/cache"
	"github.com/lazywoo/mercury/internal/interactive/repository/dao"
	"github.com/lazywoo/mercury/internal/interactive/service"
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

func InitAPP() *App {
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
