//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/article/events"
	"github.com/lazywoo/mercury/article/grpc"
	"github.com/lazywoo/mercury/article/ioc"
	"github.com/lazywoo/mercury/article/repository"
	"github.com/lazywoo/mercury/article/repository/cache"
	"github.com/lazywoo/mercury/article/repository/dao"
	"github.com/lazywoo/mercury/article/service"
	"github.com/lazywoo/mercury/pkg/wego"
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

func InitAPP() *wego.App {
	wire.Build(
		thirdProviderSet,
		svcProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
