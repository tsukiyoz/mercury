package ioc

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/tsukaychan/mercury/internal/web"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"github.com/tsukaychan/mercury/internal/web/middleware"
	"github.com/tsukaychan/mercury/pkg/ginx"
	ginxlogger "github.com/tsukaychan/mercury/pkg/ginx/middleware/logger"
	"github.com/tsukaychan/mercury/pkg/ginx/middleware/metrics"
	ginRatelimit "github.com/tsukaychan/mercury/pkg/ginx/middleware/ratelimit"
	"github.com/tsukaychan/mercury/pkg/logger"
	"github.com/tsukaychan/mercury/pkg/ratelimit"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, oAuth2Hdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler, commentHdl *web.CommentHandler, logger logger.Logger) *gin.Engine {
	ginx.SetLogger(logger)
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oAuth2Hdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	commentHdl.RegisterRoutes(server)
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
		Subsystem:  "mercury",
		Name:       "gin_http",
		Help:       "metrics gin http interface",
		InstanceID: "instance_id",
	}
	ginx.InitCounterVec(prometheus.CounterOpts{
		Namespace: "tsukiyo",
		Subsystem: "mercury",
		Name:      "biz_code",
		Help:      "metrics http biz code",
	})
	return []gin.HandlerFunc{
		corsHdl(),
		accessLogBdr.Build(),
		metricsBdr.Build(),
		otelgin.Middleware("mercury"),
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
