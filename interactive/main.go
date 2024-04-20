package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	initViper()
	initLogger()

	app := InitApp()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	go func() {
		err := app.web.Start()
		log.Println(err)
	}()

	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	cfile := pflag.String("config", "config/dev.yaml", "set config file path")
	pflag.Parse()

	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println(in.Name, in.Op)
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	zap.L().Info("Logger initialized :)")
}
