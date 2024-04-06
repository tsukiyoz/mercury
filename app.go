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

	"github.com/robfig/cron/v3"

	"github.com/gin-gonic/gin"
	"github.com/tsukaychan/mercury/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
	cron      *cron.Cron
}

func (app *App) Start(addr string) {
	for _, consumer := range app.consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}
	}

	app.cron.Start()

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

	// cron
	log.Println("shutting down cron jobs...")
	tm := time.NewTimer(time.Minute * 10)
	select {
	case <-tm.C:
		log.Println("cron jobs forced to shutdown: time expired")
	case <-app.cron.Stop().Done():
	}

	// web
	log.Println("shutting down web server...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("web server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
