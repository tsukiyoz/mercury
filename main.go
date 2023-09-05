/**
 * @author tsukiyo
 * @date 2023-08-06 12:41
 */

package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	redis_session "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"webook/config"
	"webook/internal/api"
	"webook/internal/api/middleware"
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/pkg/middleware/ratelimit"
)

func main() {
	db := initDB()
	r := initServer()
	//r := gin.Default()
	u := initUser(db)
	u.RegisterRoutes(r)
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world!")
	})
	startServer(r, ":8081")
}

func initUser(db *gorm.DB) *api.UserHandler {
	ud := dao.NewUserDao(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	uh := api.NewHandler(us)
	return uh
}

func initServer() *gin.Engine {
	r := gin.Default()

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})

	r.Use(ratelimit.NewBuilder(redisClient, time.Second, 250).Build())
	r.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://124.70.190.134")
		},
		ExposeHeaders: []string{"x-jwt-token"},
		MaxAge:        20 * time.Second,
	}))

	store, err := redis_session.NewStore(12, "tcp", config.Config.Redis.Addr, "", []byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S"), []byte("qG3mAvjIqTl2X9Hh75qaIpQg9nHU2zJf"))
	if err != nil {
		panic(err)
	}

	r.Use(sessions.Sessions("ssid", store))

	r.Use(middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths("/user/signup", "/user/login", "/hello").Build())
	return r
}

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

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
