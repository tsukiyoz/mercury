//go:build wireinject

package main

import (
	"github.com/tsukaychan/webook/internal/api"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/repository"
	captchacache "github.com/tsukaychan/webook/internal/repository/cache/captcha"
	usercache "github.com/tsukaychan/webook/internal/repository/cache/user"
	"github.com/tsukaychan/webook/internal/repository/dao"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLimiter,
		ioc.InitLogger,

		dao.NewGORMUserDAO,
		articleDao.NewGORMArticleDAO,

		usercache.NewUserRedisCache,
		captchacache.NewCaptchaRedisCache,

		repository.NewCachedUserRepository,
		repository.NewCachedCaptchaRepository,
		repository.NewCachedArticleRepository,

		service.NewUserService,
		//ioc.InitUserService,
		service.NewCaptchaService,
		service.NewArticleService,
		ioc.InitSMSService,
		ioc.InitWechatService,
		ioc.NewWechatHandlerConfig,

		api.NewUserHandler,
		api.NewOAuth2Handler,
		api.NewArticleHandler,
		ijwt.NewRedisJWTHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		//ioc.InitLocalCache,
	)
	return new(gin.Engine)
}
