package ioc

import (
	"github.com/tsukaychan/mercury/pkg/logger"
	"github.com/tsukaychan/mercury/user/repository"
	"github.com/tsukaychan/mercury/user/service"
	"go.uber.org/zap"
)

func InitUserService(repo repository.UserRepository) service.UserService {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return service.NewUserService(repo, logger.NewZapLogger(zapLogger))
}
