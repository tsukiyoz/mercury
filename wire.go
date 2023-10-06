//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/api"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,

		dao.NewUserGormDao,

		cache.NewUserRedisCache,
		cache.NewRedisCaptchaCache,

		repository.NewCachedUserRepository,
		repository.NewCachedCaptchaRepository,

		service.NewUserServiceV1,
		service.NewCaptchaServiceV1,
		ioc.InitSMSService,

		api.NewUserHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
