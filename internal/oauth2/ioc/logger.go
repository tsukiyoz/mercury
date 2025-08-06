package ioc

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/tsukiyo/mercury/pkg/logger"
)

func InitLogger() logger.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
