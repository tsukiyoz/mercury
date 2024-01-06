//go:build wireinject

package main

import (
	"webook/internal/api"
	ijwt "webook/internal/api/jwt"
	"webook/internal/repository"
	captchacache "webook/internal/repository/cache/captcha"
	usercache "webook/internal/repository/cache/user"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLimiter,
		ioc.InitLogger,

		dao.NewUserGormDao,

		usercache.NewUserRedisCache,
		captchacache.NewCaptchaRedisCache,

		repository.NewUserCachedRepository,
		repository.NewCaptchaCachedRepository,

		service.NewUserServiceV1,
		//ioc.InitUserService,
		service.NewCaptchaServiceV1,
		ioc.InitSMSService,
		ioc.InitWechatService,
		ioc.NewWechatHandlerConfig,

		api.NewUserHandler,
		api.NewOAuth2Handler,
		ijwt.NewRedisJWTHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		//ioc.InitLocalCache,
	)
	return new(gin.Engine)
}
