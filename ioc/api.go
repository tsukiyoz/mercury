package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/api"
	"webook/internal/api/middleware"
	ginRatelimit "webook/pkg/gin/middleware/ratelimit"
	"webook/pkg/ratelimit"
)

func InitWebServer(mdls []gin.HandlerFunc, hdl *api.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	return server
}

func InitLimiter(cmd redis.Cmdable) ratelimit.Limiter {
	r := ratelimit.NewRedisSlidingWindowLimiter(cmd, time.Second, 200)
	return r
}

func InitMiddlewares(limiter ratelimit.Limiter) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths(
			"/users/signup",
			"/users/login",
			"/",
			"/users/login_sms/captcha/send",
			"/users/login_sms/captcha/validate",
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
		ExposeHeaders: []string{"x-jwt-token"},
		MaxAge:        20 * time.Second,
	})
}
