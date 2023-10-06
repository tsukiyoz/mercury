/**
 * @author tsukiyo
 * @date 2023-08-06 12:41
 */

package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := InitWebServer()
	server.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "welcome to tsukiyo's website!")
	})
	startServer(server, ":8081")
}

//func initUser(r *gin.Engine, db *gorm.DB, rdb redis.Cmdable) {
//	userDao := dao.NewUserDao(db)
//	userCache := cache.NewUserCache(rdb)
//	userRepo := repository.NewUserRepository(userDao, userCache)
//	userService := service.NewUserService(userRepo)
//	captchaCache := cache.NewCaptchaCache(rdb)
//	captchaRepo := repository.NewCaptchaRepository(captchaCache)
//	smsService := memory.NewService()
//	captchaService := service.NewCaptchaService(captchaRepo, smsService)
//	uh := api.NewUserHandler(userService, captchaService)
//	uh.RegisterRoutes(r)
//}

//func initServer() *gin.Engine {
//	r := gin.Default()
//
//	//redisClient := redis.NewClient(&redis.Options{
//	//	Addr: config.Config.Redis.Addr,
//	//})
//
//	//r.Use(ratelimit.NewBuilder(redisClient, time.Second, 180).Build())
//	//r.Use(cors.New(cors.Config{
//	//	//AllowOrigins:     []string{"http://localhost:3000"},
//	//	AllowHeaders:     []string{"Content-Type", "Authorization"},
//	//	AllowCredentials: true,
//	//	AllowOriginFunc: func(origin string) bool {
//	//		return strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://124.70.190.134") || strings.HasSuffix(origin, "tsukiyo.top")
//	//	},
//	//	ExposeHeaders: []string{"x-jwt-token"},
//	//	MaxAge:        20 * time.Second,
//	//}))
//
//	store, err := redis_session.NewStore(12, "tcp", config.Config.Redis.Addr, "", []byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S"), []byte("qG3mAvjIqTl2X9Hh75qaIpQg9nHU2zJf"))
//	if err != nil {
//		panic(err)
//	}
//
//	r.Use(sessions.Sessions("ssid", store))
//
//	//r.Use(middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths("/user/signup", "/user/login", "/", "/user/login_sms/captcha/send", "/user/login_sms/captcha/validate").Build())
//	return r
//}

func startServer(r *gin.Engine, addr string) {
	fmt.Println("server started at ", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
