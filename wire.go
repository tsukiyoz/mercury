//go:build wireinject

package main

import (
	"github.com/google/wire"
	repository5 "github.com/tsukaychan/mercury/article/repository"
	articleCache "github.com/tsukaychan/mercury/article/repository/cache"
	articleDao "github.com/tsukaychan/mercury/article/repository/dao"
	service5 "github.com/tsukaychan/mercury/article/service"
	repository4 "github.com/tsukaychan/mercury/captcha/repository"
	captchaCache "github.com/tsukaychan/mercury/captcha/repository/cache"
	service4 "github.com/tsukaychan/mercury/captcha/service"
	"github.com/tsukaychan/mercury/interactive/events"
	repository2 "github.com/tsukaychan/mercury/interactive/repository"
	"github.com/tsukaychan/mercury/interactive/repository/cache"
	dao2 "github.com/tsukaychan/mercury/interactive/repository/dao"
	service2 "github.com/tsukaychan/mercury/interactive/service"
	events2 "github.com/tsukaychan/mercury/internal/events"
	"github.com/tsukaychan/mercury/internal/web"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"github.com/tsukaychan/mercury/ioc"
	"github.com/tsukaychan/mercury/ranking/repository"
	cache2 "github.com/tsukaychan/mercury/ranking/repository/cache"
	"github.com/tsukaychan/mercury/ranking/service"
	repository3 "github.com/tsukaychan/mercury/user/repository"
	userCache "github.com/tsukaychan/mercury/user/repository/cache"
	"github.com/tsukaychan/mercury/user/repository/dao"
	service3 "github.com/tsukaychan/mercury/user/service"
)

var userSvcProvider = wire.NewSet(
	service3.NewUserService,
	repository3.NewCachedUserRepository,
	dao.NewGORMUserDAO,
	userCache.NewUserRedisCache,
)

var captchaSvcProvider = wire.NewSet(
	service4.NewCaptchaService,
	captchaCache.NewCaptchaRedisCache,
	repository4.NewCachedCaptchaRepository,
)

var articleSvcProvider = wire.NewSet(
	service5.NewArticleService,
	repository5.NewCachedArticleRepository,
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
	cache2.NewRankingRedisCache,
	cache2.NewRankingLocalCache,
)

func InitAPP() *App {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLimiter,
		ioc.InitEtcdClient,
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

		web.NewCommentHandler,
		web.NewUserHandler,
		web.NewOAuth2Handler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitInteractiveClient,
		ioc.InitCommentClient,
		ioc.InitWebServer,
		ioc.InitMiddlewares,
		// ioc.InitLocalCache,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
