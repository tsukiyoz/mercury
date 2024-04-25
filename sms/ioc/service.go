package ioc

import (
	"github.com/tsukaychan/mercury/pkg/logger"
	"github.com/tsukaychan/mercury/sms/service"
	"github.com/tsukaychan/mercury/sms/service/memory"
)

func InitService(l logger.Logger) service.Service {
	return service.NewService(memory.NewService(), l)
}
