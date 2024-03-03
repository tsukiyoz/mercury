/**
 * @author tsukiyo
 * @date 2023-08-06 12:41
 */

package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {
	initViper()
	initLogger()
	initPrometheus()

	app := InitWebServer()
	app.Start()
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

func initViperRemote() {
	viper.SetConfigType("yaml")
	if err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/webook"); err != nil {
		panic(err)
	}
	if err := viper.ReadRemoteConfig(); err != nil {
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
