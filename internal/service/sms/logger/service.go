package logger

import (
	"context"
	"go.uber.org/zap"
	"webook/internal/service/sms"
)

type Service struct {
	svc sms.Service
}

func (s *Service) Send(ctx context.Context, biz string, args []sms.ArgVal, phones ...string) error {
	zap.L().Debug("send captcha", zap.String("biz", biz), zap.Any("args", args))
	err := s.svc.Send(ctx, biz, args, phones...)
	if err != nil {
		zap.L().Debug("send captcha failed", zap.Error(err))
	}
	return err
}
