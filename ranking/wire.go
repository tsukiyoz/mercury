//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/mercury/pkg/wego"
	"github.com/tsukaychan/mercury/ranking/grpc"
	"github.com/tsukaychan/mercury/ranking/ioc"
	"github.com/tsukaychan/mercury/ranking/repository"
	"github.com/tsukaychan/mercury/ranking/repository/cache"
	"github.com/tsukaychan/mercury/ranking/service"
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

func InitAPP() *wego.App {
	wire.Build(
		thirdProviderSet,
		svcProviderSet,
		grpc.NewRankingServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
