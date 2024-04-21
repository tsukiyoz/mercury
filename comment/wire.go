//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/tsukaychan/mercury/comment/grpc"
	"github.com/tsukaychan/mercury/comment/ioc"
	"github.com/tsukaychan/mercury/comment/repository"
	"github.com/tsukaychan/mercury/comment/repository/dao"
	"github.com/tsukaychan/mercury/comment/service"
)

var thirdProviderSet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
)

var serviceProviderSet = wire.NewSet(
	grpc.NewCommentServiceServer,
	service.NewCommentService,
	repository.NewCommentRepository,
	dao.NewCommentDAO,
)

func InitAPP() *App {
	wire.Build(
		thirdProviderSet,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
