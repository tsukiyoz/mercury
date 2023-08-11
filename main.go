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
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
)

func main() {
	db := initDB()

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

	u := initUser(db)

	u.RegisterRoutes(r)

	startServer(r, ":8080")
}

func initUser(db *gorm.DB) *api.UserHandler {
	ud := dao.NewUserDao(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := api.NewHandler(svc)
	return u
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:for.nothing@tcp(localhost:3306)/webook"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
