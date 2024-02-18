//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/tsukaychan/webook/internal/api"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/repository"
	articleCache "github.com/tsukaychan/webook/internal/repository/cache/article"
	captchaCache "github.com/tsukaychan/webook/internal/repository/cache/captcha"
	userCache "github.com/tsukaychan/webook/internal/repository/cache/user"
	"github.com/tsukaychan/webook/internal/repository/dao"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/ioc"
)

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitLog)
var userSvcProvider = wire.NewSet(
	dao.NewGORMUserDAO,
	userCache.NewUserRedisCache,
	repository.NewCachedUserRepository,
	service.NewUserService)
var articleSvcProvider = wire.NewSet(
	service.NewArticleService,
	repository.NewCachedArticleRepository,
	articleDao.NewGORMArticleDAO,
	articleCache.NewRedisArticleCache,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		articleSvcProvider,

		captchaCache.NewCaptchaRedisCache,
		repository.NewCachedCaptchaRepository,

		// service
		ioc.InitSMSService,
		InitPhantomWechatService,
		service.NewCaptchaService,

		// handler
		api.NewUserHandler,
		api.NewOAuth2Handler,
		api.NewArticleHandler,
		InitWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,

		// gin middleware
		ioc.InitMiddlewares,
		ioc.InitLimiter,

		// Web server
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(dao articleDao.ArticleDAO) *api.ArticleHandler {
	wire.Build(
		thirdProvider,
		service.NewArticleService,
		repository.NewCachedArticleRepository,
		articleCache.NewRedisArticleCache,
		api.NewArticleHandler,
	)
	return &api.ArticleHandler{}
}

func InitUserSvc() service.UserService {
	wire.Build(
		thirdProvider,
		userSvcProvider,
	)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	//wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}
