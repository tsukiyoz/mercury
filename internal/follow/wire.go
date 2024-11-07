//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/internal/follow/grpc"
	"github.com/lazywoo/mercury/internal/follow/ioc"
	"github.com/lazywoo/mercury/internal/follow/repository"
	"github.com/lazywoo/mercury/internal/follow/repository/cache"
	"github.com/lazywoo/mercury/internal/follow/repository/dao"
	"github.com/lazywoo/mercury/internal/follow/service"
	"github.com/lazywoo/mercury/pkg/wego"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitRedis,
)

var svcProviderSet = wire.NewSet(
	grpc.NewFollowServiceServer,
	service.NewFollowService,
	repository.NewCachedFollowRepository,
	dao.NewGORMFollowDAO,
	cache.NewRedisFollowCache,
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
