package ioc

import (
	"github.com/lazywoo/mercury/internal/web"
	wechat2 "github.com/lazywoo/mercury/oauth2/service/wechat"
	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/spf13/viper"
)

func InitWechatService(logger logger.Logger) wechat2.Service {
	type Config struct {
		AppID     string
		AppSecret string
	}

	var cfg Config
	err := viper.UnmarshalKey("wechat", &cfg)
	if err != nil {
		panic(err)
	}
	return wechat2.NewService(cfg.AppID, cfg.AppSecret, logger)
}

func NewWechatHandlerConfig() web.WechatHandlerConfig {
	type Config struct {
		Secure   bool `yaml:"secure"`
		HTTPOnly bool `yaml:"http_only"`
	}
	var cfg Config
	err := viper.UnmarshalKey("http", &cfg)
	if err != nil {
		panic(err)
	}
	return web.WechatHandlerConfig{
		Secure:   cfg.Secure,
		HTTPOnly: cfg.HTTPOnly,
	}
}
