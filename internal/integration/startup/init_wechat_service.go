package startup

import (
	"github.com/tsukaychan/webook/internal/service/oauth2/wechat"
	"github.com/tsukaychan/webook/pkg/logger"
)

func InitPhantomWechatService(l logger.Logger) wechat.Service {
	return wechat.NewService("", "", l)
}
