package ioc

import (
	"context"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/tsukaychan/webook/internal/api"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/api/middleware"
	"github.com/tsukaychan/webook/pkg/ginx"
	ginxlogger "github.com/tsukaychan/webook/pkg/ginx/middleware/logger"
	ginRatelimit "github.com/tsukaychan/webook/pkg/ginx/middleware/ratelimit"
	"github.com/tsukaychan/webook/pkg/logger"
	"github.com/tsukaychan/webook/pkg/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *api.UserHandler, oAuth2Hdl *api.OAuth2WechatHandler, articleHdl *api.ArticleHandler, logger logger.Logger) *gin.Engine {
	ginx.SetLogger(logger)
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oAuth2Hdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	return server
}

func InitLimiter(cmd redis.Cmdable) ratelimit.Limiter {
	r := ratelimit.NewRedisSlidingWindowLimiter(cmd, time.Second, 120)
	return r
}

func InitMiddlewares(limiter ratelimit.Limiter, l logger.Logger, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	bd := ginxlogger.NewMiddlewareBuilder(func(ctx context.Context, aL *ginxlogger.AccessLog) {
		l.Debug("[HTTP request]", logger.Field{
			Key:   "accessLog",
			Value: aL,
		})
	}).AllowReqBody(viper.GetBool("web.log.req")).AllowRespBody(viper.GetBool("web.log.resp"))
	viper.OnConfigChange(func(in fsnotify.Event) {
		bd.AllowReqBody(viper.GetBool("web.log.req"))
		bd.AllowRespBody(viper.GetBool("web.log.resp"))
	})
	return []gin.HandlerFunc{
		corsHdl(),
		bd.Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).IgnorePaths(
			"/",
			"/users/signup",
			"/users/login",
			"/users/refresh_token",
			"/users/login_sms/captcha/send",
			"/users/login_sms",
			"/oauth2/wechat/authurl",
			"/oauth2/wechat/callback",
		).Build(),
		ginRatelimit.NewBuilder(limiter).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://124.70.190.134") || strings.HasSuffix(origin, "tsukiyo.top")
		},
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		MaxAge:        20 * time.Second,
	})
}
