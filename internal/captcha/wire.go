//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lazywoo/mercury/internal/captcha/grpc"
	"github.com/lazywoo/mercury/internal/captcha/ioc"
	"github.com/lazywoo/mercury/internal/captcha/repository"
	"github.com/lazywoo/mercury/internal/captcha/repository/cache"
	"github.com/lazywoo/mercury/internal/captcha/service"
	"github.com/lazywoo/mercury/pkg/app"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitEtcdClient,
	ioc.InitSmsServiceClient,
)

var svcProviderSet = wire.NewSet(
	grpc.NewCaptchaServiceServer,
	service.NewCaptchaService,
	repository.NewCachedCaptchaRepository,
	cache.NewCaptchaRedisCache,
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
