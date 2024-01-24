package ioc

import (
	"github.com/tsukaychan/webook/internal/repository"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/pkg/logger"
	"go.uber.org/zap"
)

func InitUserService(repo repository.UserRepository) service.UserService {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return service.NewUserService(repo, logger.NewZapLogger(zapLogger))
}
