//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/tsukiyo/mercury/internal/article/events"
	"github.com/tsukiyo/mercury/internal/article/grpc"
	"github.com/tsukiyo/mercury/internal/article/ioc"
	"github.com/tsukiyo/mercury/internal/article/repository"
	"github.com/tsukiyo/mercury/internal/article/repository/cache"
	"github.com/tsukiyo/mercury/internal/article/repository/dao"
	"github.com/tsukiyo/mercury/internal/article/service"
	"github.com/tsukiyo/mercury/pkg/app"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.NewSyncProducer,
	ioc.InitEtcdClient,
	ioc.InitUserRpcClient,
)

var svcProviderSet = wire.NewSet(
	grpc.NewArticleServiceServer,
	events.NewSaramaSyncProducer,
	service.NewArticleService,
	repository.NewCachedArticleRepository,
	dao.NewGORMArticleDAO,
	cache.NewRedisArticleCache,
)

func InitAPP() *app.App {
	wire.Build(
		thirdProviderSet,
		svcProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(app.App), "GRPCServer"),
	)
	return new(app.App)
}
