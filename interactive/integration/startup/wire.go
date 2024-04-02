//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/tsukaychan/webook/interactive/grpc"
	repository2 "github.com/tsukaychan/webook/interactive/repository"
	"github.com/tsukaychan/webook/interactive/repository/cache"
	dao2 "github.com/tsukaychan/webook/interactive/repository/dao"
	service2 "github.com/tsukaychan/webook/interactive/service"
)

var thirdProvider = wire.NewSet(
	InitRedis,
	InitTestDB,
	InitLog,
	InitKafka,
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service2.NewInteractiveService(nil, nil)
}

func InitInteractiveGRPCServer() *grpc.InteractiveServiceServer {
	wire.Build(thirdProvider, interactiveSvcProvider, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
