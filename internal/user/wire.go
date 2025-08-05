//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/tsukiyo/mercury/internal/user/grpc"
	"github.com/tsukiyo/mercury/internal/user/ioc"
	"github.com/tsukiyo/mercury/internal/user/repository"
	"github.com/tsukiyo/mercury/internal/user/repository/cache"
	"github.com/tsukiyo/mercury/internal/user/repository/dao"
	"github.com/tsukiyo/mercury/internal/user/service"
	"github.com/tsukiyo/mercury/pkg/app"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitEtcdClient,
)

var svcProviderSet = wire.NewSet(
	grpc.NewUserServiceServer,
	service.NewUserService,
	repository.NewCachedUserRepository,
	dao.NewGORMUserDAO,
	cache.NewUserRedisCache,
)

func InitAPP() *app.App {
	wire.Build(
		thirdProviderSet,
		ioc.InitGRPCxServer,
		svcProviderSet,
		wire.Struct(new(app.App), "GRPCServer"),
	)
	return new(app.App)
}
