package ioc

import (
	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/lazywoo/mercury/user/repository"
	"github.com/lazywoo/mercury/user/service"
	"go.uber.org/zap"
)

func InitUserService(repo repository.UserRepository) service.UserService {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return service.NewUserService(repo, logger.NewZapLogger(zapLogger))
}
