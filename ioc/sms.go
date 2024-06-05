package ioc

import (
	"github.com/lazywoo/mercury/internal/service/sms"
	"github.com/lazywoo/mercury/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
