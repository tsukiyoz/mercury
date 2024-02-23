//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/tsukaychan/webook/internal/api"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/repository"
	articleCache "github.com/tsukaychan/webook/internal/repository/cache/article"
	captchacache "github.com/tsukaychan/webook/internal/repository/cache/captcha"
	cache "github.com/tsukaychan/webook/internal/repository/cache/interactive"
	usercache "github.com/tsukaychan/webook/internal/repository/cache/user"
	"github.com/tsukaychan/webook/internal/repository/dao"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLimiter,
		ioc.InitLogger,

		dao.NewGORMUserDAO,
		articleDao.NewGORMArticleDAO,
		dao.NewGORMInteractiveDAO,

		usercache.NewUserRedisCache,
		captchacache.NewCaptchaRedisCache,
		articleCache.NewRedisArticleCache,
		cache.NewRedisInteractiveCache,

		repository.NewCachedUserRepository,
		repository.NewCachedCaptchaRepository,
		repository.NewCachedArticleRepository,
		repository.NewCachedInteractiveRepository,

		service.NewUserService,
		// ioc.InitUserService,
		service.NewCaptchaService,
		service.NewArticleService,
		service.NewInteractiveService,
		ioc.InitSMSService,
		ioc.InitWechatService,
		ioc.NewWechatHandlerConfig,

		api.NewUserHandler,
		api.NewOAuth2Handler,
		api.NewArticleHandler,
		ijwt.NewRedisJWTHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		// ioc.InitLocalCache,
	)
	return new(gin.Engine)
}
