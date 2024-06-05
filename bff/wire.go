//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/bff/ioc"
	"github.com/lazywoo/mercury/bff/web"
	"github.com/lazywoo/mercury/bff/web/jwt"
	"github.com/lazywoo/mercury/pkg/wego"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitEtcdClient,
)

var hdlProviderSet = wire.NewSet(
	web.NewUserHandler,
	jwt.NewRedisJWTHandler,
	web.NewOAuth2Handler,
	web.NewArticleHandler,
	web.NewCommentHandler,
)

var cliProviderSet = wire.NewSet(
	ioc.InitUserClient,
	ioc.InitCaptchaClient,
	ioc.InitOAuth2Client,
	ioc.InitArticleClient,
	ioc.InitInteractiveClient,
	ioc.InitCommentClient,
)

func InitAPP() *wego.App {
	wire.Build(
		thirdProviderSet,
		hdlProviderSet,
		cliProviderSet,
		ioc.InitWebLimiter,
		ioc.InitWebServer,
		ioc.InitWechatHandlerConfig,
		wire.Struct(new(wego.App), "WebServer"),
	)
	return new(wego.App)
}
