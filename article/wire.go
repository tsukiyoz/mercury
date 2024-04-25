//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/mercury/article/events"
	"github.com/tsukaychan/mercury/article/grpc"
	"github.com/tsukaychan/mercury/article/ioc"
	"github.com/tsukaychan/mercury/article/repository"
	"github.com/tsukaychan/mercury/article/repository/cache"
	"github.com/tsukaychan/mercury/article/repository/dao"
	"github.com/tsukaychan/mercury/article/service"
	"github.com/tsukaychan/mercury/pkg/wego"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.NewSyncProducer,
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
