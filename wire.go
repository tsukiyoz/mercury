//go:build wireinject

package main

import (
	"github.com/google/wire"
	events2 "github.com/tsukaychan/webook/internal/events"
	"github.com/tsukaychan/webook/internal/repository"
	articleCache "github.com/tsukaychan/webook/internal/repository/cache/article"
	captchacache "github.com/tsukaychan/webook/internal/repository/cache/captcha"
	cache "github.com/tsukaychan/webook/internal/repository/cache/interactive"
	usercache "github.com/tsukaychan/webook/internal/repository/cache/user"
	"github.com/tsukaychan/webook/internal/repository/dao"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/internal/web"
	ijwt "github.com/tsukaychan/webook/internal/web/jwt"
	"github.com/tsukaychan/webook/ioc"
)

func InitWebServer() *App {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLimiter,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewSyncProducer,
		ioc.NewConsumers,

		events2.NewInteractiveReadEventConsumer,
		events2.NewSaramaSyncProducer,

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

		web.NewUserHandler,
		web.NewOAuth2Handler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		// ioc.InitLocalCache,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
