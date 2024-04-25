package startup

import (
	wechat2 "github.com/tsukaychan/mercury/oauth2/service/wechat"
	"github.com/tsukaychan/mercury/pkg/logger"
)

func InitPhantomWechatService(l logger.Logger) wechat2.Service {
	return wechat2.NewService("", "", l)
}
