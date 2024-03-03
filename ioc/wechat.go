package ioc

import (
	"github.com/spf13/viper"
	"github.com/tsukaychan/webook/internal/service/oauth2/wechat"
	"github.com/tsukaychan/webook/internal/web"
	"github.com/tsukaychan/webook/pkg/logger"
)

func InitWechatService(logger logger.Logger) wechat.Service {
	type Config struct {
		AppID     string
		AppSecret string
	}

	var cfg Config
	err := viper.UnmarshalKey("wechat", &cfg)
	if err != nil {
		panic(err)
	}
	return wechat.NewService(cfg.AppID, cfg.AppSecret, logger)
}

func NewWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}
