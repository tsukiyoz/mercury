//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/api"
	"webook/internal/repository"
	"webook/internal/repository/cache/captcha"
	"webook/internal/repository/cache/user"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,

		dao.NewUserGormDao,

		user.NewUserRedisCache,
		captcha.NewCaptchaLocalCache,

		repository.NewCachedUserRepository,
		repository.NewCachedCaptchaRepository,

		service.NewUserServiceV1,
		service.NewCaptchaServiceV1,
		ioc.InitSMSService,

		api.NewUserHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		ioc.InitLocalCache,
	)
	return new(gin.Engine)
}
