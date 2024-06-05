//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/pkg/wego"
	"github.com/lazywoo/mercury/user/grpc"
	"github.com/lazywoo/mercury/user/ioc"
	"github.com/lazywoo/mercury/user/repository"
	"github.com/lazywoo/mercury/user/repository/cache"
	"github.com/lazywoo/mercury/user/repository/dao"
	"github.com/lazywoo/mercury/user/service"
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

func InitAPP() *wego.App {
	wire.Build(
		thirdProviderSet,
		ioc.InitGRPCxServer,
		svcProviderSet,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
