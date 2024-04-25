//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/tsukaychan/mercury/article/repository"
	articleCache "github.com/tsukaychan/mercury/article/repository/cache"
	dao3 "github.com/tsukaychan/mercury/article/repository/dao"
	"github.com/tsukaychan/mercury/article/service"
	repository4 "github.com/tsukaychan/mercury/captcha/repository"
	captchaCache "github.com/tsukaychan/mercury/captcha/repository/cache"
	service4 "github.com/tsukaychan/mercury/captcha/service"
	repository3 "github.com/tsukaychan/mercury/interactive/repository"
	"github.com/tsukaychan/mercury/interactive/repository/cache"
	dao2 "github.com/tsukaychan/mercury/interactive/repository/dao"
	service2 "github.com/tsukaychan/mercury/interactive/service"
	"github.com/tsukaychan/mercury/internal/events"
	"github.com/tsukaychan/mercury/internal/web"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"github.com/tsukaychan/mercury/ioc"
	repository2 "github.com/tsukaychan/mercury/user/repository"
	userCache "github.com/tsukaychan/mercury/user/repository/cache"
	"github.com/tsukaychan/mercury/user/repository/dao"
	service3 "github.com/tsukaychan/mercury/user/service"
)

var thirdProvider = wire.NewSet(
	InitRedis,
	InitTestDB,
	InitLog,
	InitKafka,
	NewSyncProducer,
)

var userSvcProvider = wire.NewSet(
	service3.NewUserService,
	events.NewSaramaSyncProducer,
	repository2.NewCachedUserRepository,
	dao.NewGORMUserDAO,
	userCache.NewUserRedisCache,
)

var articleSvcProvider = wire.NewSet(
	service.NewArticleService,
	repository.NewCachedArticleRepository,
	dao3.NewGORMArticleDAO,
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

		service4.NewCaptchaService,
		repository4.NewCachedCaptchaRepository,
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

func InitArticleHandler(atclDao dao3.ArticleDAO) *web.ArticleHandler {
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

func InitUserSvc() service3.UserService {
	wire.Build(
		thirdProvider,
		userSvcProvider,
	)
	return service3.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	// wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}
