//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/tsukiyo/mercury/internal/bff/ioc"
	"github.com/tsukiyo/mercury/internal/bff/web"
	"github.com/tsukiyo/mercury/internal/bff/web/jwt"
	"github.com/tsukiyo/mercury/pkg/app"
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

func InitAPP() *app.App {
	wire.Build(
		thirdProviderSet,
		hdlProviderSet,
		cliProviderSet,
		ioc.InitWebLimiter,
		ioc.InitWebServer,
		ioc.InitWechatHandlerConfig,
		wire.Struct(new(app.App), "WebServer"),
	)
	return new(app.App)
}
