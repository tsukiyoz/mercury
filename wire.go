//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/api"
	"webook/internal/repository"
	captchacache "webook/internal/repository/cache/captcha"
	usercache "webook/internal/repository/cache/user"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLimiter,

		dao.NewUserGormDao,

		usercache.NewUserRedisCache,
		captchacache.NewCaptchaRedisCache,

		repository.NewUserCachedRepository,
		repository.NewCaptchaCachedRepository,

		service.NewUserServiceV1,
		service.NewCaptchaServiceV1,
		ioc.InitSMSService,

		api.NewUserHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		//ioc.InitLocalCache,
	)
	return new(gin.Engine)
}
