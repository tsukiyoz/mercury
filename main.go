/**
 * @author tsukiyo
 * @date 2023-08-06 12:41
 */

package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"webook/internal/api"
	"webook/internal/api/middleware"
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
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
	startServer(r, ":8080")
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
	r.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://124.70.190.134")
		},
		MaxAge: 20 * time.Second,
	}))

	//store := cookie.NewStore([]byte("secret"))
	//store := memstore.NewStore([]byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S"), []byte("qG3mAvjIqTl2X9Hh75qaIpQg9nHU2zJf"))
	//newStore, err := redis.NewStore(12, "tcp", "localhost:6379", "", []byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S"), []byte("qG3mAvjIqTl2X9Hh75qaIpQg9nHU2zJf"))
	//if err != nil {
	//	panic(err)
	//}
	//store := newStore
	//r.Use(sessions.Sessions("mysession", store))

	r.Use(middleware.NewLoginMiddlewareBuilder().IgnorePaths("/user/signup", "/user/login").Build())
	return r
}

func startServer(r *gin.Engine, addr string) {
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
	db, err := gorm.Open(mysql.Open("root:for.nothing@tcp(mysql-service:3309)/webook"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
