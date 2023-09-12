package retryable

import (
	"context"
	"webook/internal/service/sms"
)

type RetryService struct {
	svc      sms.Service
	retryCnt int
}

func (s *RetryService) Send(ctx context.Context, tpl string, args []sms.ArgVal, phones ...string) error {
	err := s.svc.Send(ctx, tpl, args, phones...)
	for err != nil && s.retryCnt < 10 {
		s.retryCnt++
		err = s.svc.Send(ctx, tpl, args, phones...)
	}
	return err
}
