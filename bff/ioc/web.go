package ioc

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/tsukaychan/mercury/pkg/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/tsukaychan/mercury/bff/web"
	"github.com/tsukaychan/mercury/bff/web/jwt"
	"github.com/tsukaychan/mercury/bff/web/middleware"
	"github.com/tsukaychan/mercury/pkg/ginx"
	ginRatelimit "github.com/tsukaychan/mercury/pkg/ginx/middleware/ratelimit"
	"github.com/tsukaychan/mercury/pkg/logger"
)

func InitWebServer(limiter ratelimit.Limiter, jwtHdl jwt.Handler, userHdl *web.UserHandler, oAuth2Hdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler, commentHdl *web.CommentHandler, logger logger.Logger) *ginx.Server {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	engine := gin.Default()
	engine.Use(
		corsHdl(),
		timeout(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).Build(),
		ginRatelimit.NewBuilder(limiter).Build(),
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

func timeout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, ok := ctx.Request.Context().Deadline()
		if !ok {
			newCtx, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
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
