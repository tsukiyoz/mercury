package async

import (
	"context"
	"github.com/tsukaychan/webook/internal/service/sms"
)

var _ sms.Service = (*SMSService)(nil)

type SMSService struct {
	svc sms.Service
	//dao repository
}

func (s *SMSService) StartAsync() {
	go func() {
		// 找到没发出去的请求
		// 将请求发出去，并且控制重试
	}()
}

func (s *SMSService) Send(ctx context.Context, biz string, args []sms.ArgVal, phones ...string) error {
	// 正常路径
	err := s.svc.Send(ctx, biz, args, phones...)
	if err != nil {
		// 判定是否崩溃
		// 转异步
	}
	return nil
}
