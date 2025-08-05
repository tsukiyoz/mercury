package ioc

import (
	"github.com/tsukiyo/mercury/internal/sms/service"
	"github.com/tsukiyo/mercury/internal/sms/service/memory"
	"github.com/tsukiyo/mercury/pkg/logger"
)

func InitService(l logger.Logger) service.Service {
	return service.NewService(memory.NewService(), l)
}
