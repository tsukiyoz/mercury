package ioc

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	ginxlogger "github.com/lazywoo/mercury/pkg/ginx/middleware/logger"
	"github.com/lazywoo/mercury/pkg/ginx/middleware/metrics"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/redis/go-redis/v9"

	"github.com/lazywoo/mercury/pkg/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lazywoo/mercury/bff/web"
	"github.com/lazywoo/mercury/bff/web/jwt"
	"github.com/lazywoo/mercury/bff/web/middleware"
	"github.com/lazywoo/mercury/pkg/ginx"
	ginRatelimit "github.com/lazywoo/mercury/pkg/ginx/middleware/ratelimit"
	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

func InitWebServer(limiter ratelimit.Limiter, jwtHdl jwt.Handler, userHdl *web.UserHandler, oAuth2Hdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler, commentHdl *web.CommentHandler, logger logger.Logger) *ginx.Server {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	engine := gin.Default()
	engine.Use(
		corsHdl(),
		timeout(),
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
		// ginRatelimit.NewBuilder(limiter).Build(),
	)
	userHdl.RegisterRoutes(engine)
	oAuth2Hdl.RegisterRoutes(engine)
	articleHdl.RegisterRoutes(engine)
	commentHdl.RegisterRoutes(engine)
	web.NewObservabilityHandler().RegisterRoutes(engine)
	addr := viper.GetString("http.addr")
	ginx.InitCounterVec(prometheus.CounterOpts{
		Namespace: "mercury",
		Subsystem: "bff",
		Name:      "http",
	})
	ginx.SetLogger(logger)
	return &ginx.Server{
		Engine: engine,
		Addr:   addr,
	}
}

func InitWebLimiter(cmd redis.Cmdable) ratelimit.Limiter {
	interval, rate := viper.GetUint("web.limit.interval"), viper.GetUint("web.limit.rate")
	r := ratelimit.NewRedisSlidingWindowLimiter(cmd, time.Duration(interval)*time.Second, int(rate))
	return r
}

func InitMiddlewares(limiter ratelimit.Limiter, l logger.Logger, jwtHdl jwt.Handler) []gin.HandlerFunc {
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

func timeout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, ok := ctx.Request.Context().Deadline()
		if !ok {
			newCtx, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*3)
			defer cancel()
			ctx.Request = ctx.Request.Clone(newCtx)
		}
		ctx.Next()
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
