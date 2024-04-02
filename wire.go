//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/webook/interactive/events"
	repository2 "github.com/tsukaychan/webook/interactive/repository"
	"github.com/tsukaychan/webook/interactive/repository/cache"
	dao2 "github.com/tsukaychan/webook/interactive/repository/dao"
	service2 "github.com/tsukaychan/webook/interactive/service"
	events2 "github.com/tsukaychan/webook/internal/events"
	"github.com/tsukaychan/webook/internal/repository"
	articleCache "github.com/tsukaychan/webook/internal/repository/cache/article"
	captchaCache "github.com/tsukaychan/webook/internal/repository/cache/captcha"
	rankingCache "github.com/tsukaychan/webook/internal/repository/cache/ranking"
	userCache "github.com/tsukaychan/webook/internal/repository/cache/user"
	"github.com/tsukaychan/webook/internal/repository/dao"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/internal/web"
	ijwt "github.com/tsukaychan/webook/internal/web/jwt"
	"github.com/tsukaychan/webook/ioc"
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
