//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/lazywoo/mercury/internal/comment/grpc"
	"github.com/lazywoo/mercury/internal/comment/ioc"
	"github.com/lazywoo/mercury/internal/comment/repository"
	"github.com/lazywoo/mercury/internal/comment/repository/dao"
	"github.com/lazywoo/mercury/internal/comment/service"
	"github.com/lazywoo/mercury/pkg/app"
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

func InitAPP() *app.App {
	wire.Build(
		thirdProviderSet,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(app.App), "GRPCServer"),
	)
	return new(app.App)
}
