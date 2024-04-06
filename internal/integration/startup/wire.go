//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	repository3 "github.com/tsukaychan/mercury/interactive/repository"
	"github.com/tsukaychan/mercury/interactive/repository/cache"
	dao2 "github.com/tsukaychan/mercury/interactive/repository/dao"
	service2 "github.com/tsukaychan/mercury/interactive/service"
	"github.com/tsukaychan/mercury/internal/events"
	"github.com/tsukaychan/mercury/internal/repository"
	articleCache "github.com/tsukaychan/mercury/internal/repository/cache/article"
	captchaCache "github.com/tsukaychan/mercury/internal/repository/cache/captcha"
	userCache "github.com/tsukaychan/mercury/internal/repository/cache/user"
	"github.com/tsukaychan/mercury/internal/repository/dao"
	articleDao "github.com/tsukaychan/mercury/internal/repository/dao/article"
	"github.com/tsukaychan/mercury/internal/service"
	"github.com/tsukaychan/mercury/internal/web"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"github.com/tsukaychan/mercury/ioc"
)

var thirdProvider = wire.NewSet(
	InitRedis,
	InitTestDB,
	InitLog,
	InitKafka,
	NewSyncProducer,
)

var userSvcProvider = wire.NewSet(
	service.NewUserService,
	events.NewSaramaSyncProducer,
	repository.NewCachedUserRepository,
	dao.NewGORMUserDAO,
	userCache.NewUserRedisCache,
)

var articleSvcProvider = wire.NewSet(
	service.NewArticleService,
	repository.NewCachedArticleRepository,
	articleDao.NewGORMArticleDAO,
	articleCache.NewRedisArticleCache,
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository3.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,

		userSvcProvider,
		articleSvcProvider,
		interactiveSvcProvider,

		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2Handler,

		service.NewCaptchaService,
		repository.NewCachedCaptchaRepository,
		captchaCache.NewCaptchaRedisCache,

		ioc.InitSMSService,
		InitPhantomWechatService,
		InitWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,

		ioc.InitMiddlewares,
		ioc.InitLimiter,

		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(atclDao articleDao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		userSvcProvider,
		service.NewArticleService,
		repository.NewCachedArticleRepository,
		articleCache.NewRedisArticleCache,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

func InitUserSvc() service.UserService {
	wire.Build(
		thirdProvider,
		userSvcProvider,
	)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	// wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}
