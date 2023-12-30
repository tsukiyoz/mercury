package ioc

import (
	"os"
	"webook/internal/api"
	"webook/internal/service/oauth2/wechat"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("no environment variables found WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("no environment variables found WECHAT_APP_SECRET")
	}
	return wechat.NewService(appId, appSecret)
}

func NewWechatHandlerConfig() api.WechatHandlerConfig {
	return api.WechatHandlerConfig{
		Secure: false,
	}
}
