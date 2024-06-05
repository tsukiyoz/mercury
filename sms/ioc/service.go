package ioc

import (
	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/lazywoo/mercury/sms/service"
	"github.com/lazywoo/mercury/sms/service/memory"
)

func InitService(l logger.Logger) service.Service {
	return service.NewService(memory.NewService(), l)
}
