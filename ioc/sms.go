package ioc

import (
	"github.com/tsukaychan/webook/internal/service/sms"
	"github.com/tsukaychan/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
