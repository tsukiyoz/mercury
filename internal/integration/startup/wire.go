//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/tsukaychan/webook/internal/api"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/repository"
	"github.com/tsukaychan/webook/internal/repository/article"
	captchacache "github.com/tsukaychan/webook/internal/repository/cache/captcha"
	usercache "github.com/tsukaychan/webook/internal/repository/cache/user"
	"github.com/tsukaychan/webook/internal/repository/dao"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/ioc"
)

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitLog)
var userSvcProvider = wire.NewSet(
	dao.NewGORMUserDAO,
	usercache.NewUserRedisCache,
	repository.NewCachedUserRepository,
	service.NewUserService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		//articlSvcProvider,
		captchacache.NewCaptchaRedisCache,
		dao.NewGORMArticleDAO,
		repository.NewCachedCaptchaRepository,
		article.NewCachedArticleRepository,

		// service
		ioc.InitSMSService,
		InitPhantomWechatService,
		service.NewCaptchaService,
		service.NewArticleService,

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

func InitArticleHandler() *api.ArticleHandler {
	wire.Build(
		thirdProvider,
		api.NewArticleHandler,
		service.NewArticleService,
		article.NewCachedArticleRepository,
		dao.NewGORMArticleDAO,
	)
	return &api.ArticleHandler{}
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	//wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}
