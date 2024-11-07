package ioc

import (
	"github.com/lazywoo/mercury/internal/sms/service"
	"github.com/lazywoo/mercury/internal/sms/service/memory"
	"github.com/lazywoo/mercury/pkg/logger"
)

func InitService(l logger.Logger) service.Service {
	return service.NewService(memory.NewService(), l)
}
