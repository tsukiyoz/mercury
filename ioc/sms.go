package ioc

import (
	"github.com/tsukaychan/mercury/internal/service/sms"
	"github.com/tsukaychan/mercury/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
