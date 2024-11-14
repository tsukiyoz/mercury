//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lazywoo/mercury/internal/payment/grpc"
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
	ioc.InitCronJobs,
	ioc.InitRedis,
	ioc.InitRLockClient,
)

func InitAPP() *app.App {
	wire.Build(
		thirdPartySet,

		dao.NewGORMPaymentDAO,
		repository.NewPaymentRepository,
		ioc.InitWechatNativeService,
		web.NewWechatHandler,
		grpc.NewWechatPaymentServiceServer,
		ioc.InitSyncWechatPaymentJob,
		ioc.InitWebServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(app.App), "WebServer", "GRPCServer", "Cron"),
	)
	return &app.App{}
}
