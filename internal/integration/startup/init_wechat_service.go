package startup

import (
	"github.com/tsukaychan/mercury/internal/service/oauth2/wechat"
	"github.com/tsukaychan/mercury/pkg/logger"
)

func InitPhantomWechatService(l logger.Logger) wechat.Service {
	return wechat.NewService("", "", l)
}
