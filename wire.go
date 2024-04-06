//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/mercury/interactive/events"
	repository2 "github.com/tsukaychan/mercury/interactive/repository"
	"github.com/tsukaychan/mercury/interactive/repository/cache"
	dao2 "github.com/tsukaychan/mercury/interactive/repository/dao"
	service2 "github.com/tsukaychan/mercury/interactive/service"
	events2 "github.com/tsukaychan/mercury/internal/events"
	"github.com/tsukaychan/mercury/internal/repository"
	articleCache "github.com/tsukaychan/mercury/internal/repository/cache/article"
	captchaCache "github.com/tsukaychan/mercury/internal/repository/cache/captcha"
	rankingCache "github.com/tsukaychan/mercury/internal/repository/cache/ranking"
	userCache "github.com/tsukaychan/mercury/internal/repository/cache/user"
	"github.com/tsukaychan/mercury/internal/repository/dao"
	articleDao "github.com/tsukaychan/mercury/internal/repository/dao/article"
	"github.com/tsukaychan/mercury/internal/service"
	"github.com/tsukaychan/mercury/internal/web"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"github.com/tsukaychan/mercury/ioc"
)

var userSvcProvider = wire.NewSet(
	service.NewUserService,
	repository.NewCachedUserRepository,
	dao.NewGORMUserDAO,
	userCache.NewUserRedisCache,
)

var captchaSvcProvider = wire.NewSet(
	service.NewCaptchaService,
	captchaCache.NewCaptchaRedisCache,
	repository.NewCachedCaptchaRepository,
)

var articleSvcProvider = wire.NewSet(
	service.NewArticleService,
	repository.NewCachedArticleRepository,
	articleDao.NewGORMArticleDAO,
	articleCache.NewRedisArticleCache,
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

var rankingSvcSet = wire.NewSet(
	service.NewBatchRankingService,
	repository.NewRankingCachedRepository,
	rankingCache.NewRankingRedisCache,
	rankingCache.NewRankingLocalCache,
)

func InitWebServer() *App {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLimiter,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewSyncProducer,
		ioc.NewConsumers,
		ioc.InitTasks,
		ioc.InitRankingJob,
		ioc.InitRLockClient,

		rankingSvcSet,
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcProvider,
		captchaSvcProvider,

		events.NewInteractiveReadEventConsumer,
		events2.NewSaramaSyncProducer,

		ioc.InitSMSService,
		ioc.InitWechatService,
		ioc.NewWechatHandlerConfig,

		web.NewUserHandler,
		web.NewOAuth2Handler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitInteractiveGRPCClient,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		// ioc.InitLocalCache,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
