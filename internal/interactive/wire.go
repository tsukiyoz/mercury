//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/tsukiyo/mercury/internal/interactive/events"
	"github.com/tsukiyo/mercury/internal/interactive/grpc"
	"github.com/tsukiyo/mercury/internal/interactive/ioc"
	"github.com/tsukiyo/mercury/internal/interactive/repository"
	"github.com/tsukiyo/mercury/internal/interactive/repository/cache"
	"github.com/tsukiyo/mercury/internal/interactive/repository/dao"
	"github.com/tsukiyo/mercury/internal/interactive/service"
	"github.com/tsukiyo/mercury/pkg/app"
)

var thirdProvider = wire.NewSet(
	ioc.InitSrcDB,
	ioc.InitDstDB,
	ioc.InitDualWritePool,
	ioc.InitDualWriteDB,
	// ioc.initDB,
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

func InitAPP() *app.App {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		migratorSet,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(app.App), "GRPCServer", "Consumers"),
	)
	return new(app.App)
}
