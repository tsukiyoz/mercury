package ioc

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/tsukiyo/mercury/pkg/logger"
)

func InitLogger() logger.Logger {
	type Config struct {
		Mode             string   `yaml:"mode"`
		Encoding         string   `yaml:"encoding"`
		OutputPaths      []string `yaml:"outputPaths"`
		ErrorOutputPaths []string `yaml:"errorOutputPaths"`
	}

	var c Config
	err := viper.UnmarshalKey("log", &c)
	if err != nil {
		panic(err)
	}

	var cfg zap.Config
	if c.Mode == "dev" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	if c.Encoding != "" {
		cfg.Encoding = c.Encoding
	}
	if len(c.OutputPaths) > 0 {
		cfg.OutputPaths = c.OutputPaths
	}
	if len(c.ErrorOutputPaths) > 0 {
		cfg.ErrorOutputPaths = c.ErrorOutputPaths
	}
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	// cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
