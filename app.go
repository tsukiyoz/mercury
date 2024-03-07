package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tsukaychan/webook/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}

func (app *App) Start(addr string) {
	for _, consumer := range app.consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}
	}

	server := app.web
	server.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "welcome to tsukiyo's website!")
	})

	app.startServer(addr)
}

func (app *App) startServer(addr string) {
	fmt.Println("server started at ", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: app.web,
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
