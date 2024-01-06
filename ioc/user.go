package ioc

import (
	"go.uber.org/zap"
	"webook/internal/repository"
	"webook/internal/service"
	"webook/pkg/logger"
)

func InitUserService(repo repository.UserRepository) service.UserService {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return service.NewUserServiceV1(repo, logger.NewZapLogger(zapLogger))
}
