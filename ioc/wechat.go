package ioc

import (
	"github.com/tsukaychan/webook/internal/api"
	"github.com/tsukaychan/webook/internal/service/oauth2/wechat"
	"github.com/tsukaychan/webook/pkg/logger"
	"os"
)

func InitWechatService(logger logger.Logger) wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("no environment variables found WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("no environment variables found WECHAT_APP_SECRET")
	}
	return wechat.NewService(appId, appSecret, logger)
}

func NewWechatHandlerConfig() api.WechatHandlerConfig {
	return api.WechatHandlerConfig{
		Secure: false,
	}
}
