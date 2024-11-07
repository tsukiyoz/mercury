//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lazywoo/mercury/internal/ranking/grpc"
	"github.com/lazywoo/mercury/internal/ranking/ioc"
	"github.com/lazywoo/mercury/internal/ranking/repository"
	"github.com/lazywoo/mercury/internal/ranking/repository/cache"
	"github.com/lazywoo/mercury/internal/ranking/service"
	"github.com/lazywoo/mercury/pkg/app"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitEtcdClient,
	ioc.InitArticleRpcClient,
	ioc.InitInteractiveRpcClient,
)

var svcProviderSet = wire.NewSet(
	service.NewBatchRankingService,
	repository.NewRankingCachedRepository,
	cache.NewRankingLocalCache,
	cache.NewRankingRedisCache,
)

var cronProviderSet = wire.NewSet(
	ioc.InitTasks,
	ioc.InitRankingJob,
	ioc.InitRLockClient,
)

func InitAPP() *app.App {
	wire.Build(
		thirdProviderSet,
		svcProviderSet,
		cronProviderSet,
		grpc.NewRankingServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(app.App), "GRPCServer", "Cron"),
	)
	return new(app.App)
}
