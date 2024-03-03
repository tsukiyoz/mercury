package ioc

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	ginxlogger "github.com/tsukaychan/webook/pkg/ginx/middleware/logger"

	"github.com/tsukaychan/webook/pkg/ginx/middleware/metrics"

	"github.com/tsukaychan/webook/internal/web"
	ijwt "github.com/tsukaychan/webook/internal/web/jwt"
	"github.com/tsukaychan/webook/internal/web/middleware"
	"github.com/tsukaychan/webook/pkg/ginx"
	ginRatelimit "github.com/tsukaychan/webook/pkg/ginx/middleware/ratelimit"
	"github.com/tsukaychan/webook/pkg/logger"
	"github.com/tsukaychan/webook/pkg/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, oAuth2Hdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler, logger logger.Logger) *gin.Engine {
	ginx.SetLogger(logger)
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oAuth2Hdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	web.NewObservabilityHandler().RegisterRoutes(server)
	return server
}

func InitLimiter(cmd redis.Cmdable) ratelimit.Limiter {
	interval, rate := viper.GetUint("web.limit.interval"), viper.GetUint("web.limit.rate")
	r := ratelimit.NewRedisSlidingWindowLimiter(cmd, time.Duration(interval)*time.Second, int(rate))
	return r
}

func InitMiddlewares(limiter ratelimit.Limiter, l logger.Logger, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	accessLogBdr := ginxlogger.NewMiddlewareBuilder(func(ctx context.Context, aL *ginxlogger.AccessLog) {
		l.Debug("[HTTP request]", logger.Field{
			Key:   "accessLog",
			Value: aL,
		})
	}).AllowReqBody(viper.GetBool("web.log.req")).AllowRespBody(viper.GetBool("web.log.resp"))
	viper.OnConfigChange(func(in fsnotify.Event) {
		accessLogBdr.AllowReqBody(viper.GetBool("web.log.req"))
		accessLogBdr.AllowRespBody(viper.GetBool("web.log.resp"))
	})
	metricsBdr := &metrics.PrometheusBuilder{
		Namespace:  "tsukiyo",
		Subsystem:  "webook",
		Name:       "gin_http",
		Help:       "metrics gin http interface",
		InstanceID: "instance_id",
	}
	return []gin.HandlerFunc{
		corsHdl(),
		accessLogBdr.Build(),
		metricsBdr.Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).IgnorePaths(
			"/",
			"/users/signup",
			"/users/login",
			"/users/refresh_token",
			"/users/login_sms/captcha/send",
			"/users/login_sms",
			"/oauth2/wechat/authurl",
			"/oauth2/wechat/callback",
			"/test/metric",
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
