//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lazywoo/mercury/internal/payment/ioc"
	"github.com/lazywoo/mercury/internal/payment/repository"
	"github.com/lazywoo/mercury/internal/payment/repository/dao"
	"github.com/lazywoo/mercury/internal/payment/web"
	"github.com/lazywoo/mercury/pkg/app"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitProducer,
	ioc.InitWechatNotifyHandler,
	ioc.InitWechatConfig,
	ioc.InitWechatClient,
)

func InitAPP() *app.App {
	wire.Build(
		thirdPartySet,

		dao.NewGORMPaymentDAO,
		repository.NewPaymentRepository,
		ioc.InitWechatNativeService,
		web.NewWechatHandler,
		ioc.InitWebServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(app.App), "WebServer", "GRPCServer"),
	)
	return &app.App{}
}
