package ioc

import (
	"github.com/spf13/viper"

	"github.com/lazywoo/mercury/internal/oauth2/service/wechat"
	"github.com/lazywoo/mercury/pkg/logger"
)

func InitWechatService(l logger.Logger) wechat.Service {
	type Config struct {
		AppID     string `yaml:"app_id"`
		AppSecret string `yaml:"app_secret"`
	}
	var cfg Config
	err := viper.UnmarshalKey("wechat", &cfg)
	if err != nil {
		panic(err)
	}
	return wechat.NewService(cfg.AppID, cfg.AppSecret, l)
}
