package retryable

import (
	"context"
	"errors"

	"github.com/tsukaychan/mercury/internal/service/sms"
)

type RetryService struct {
	svc      sms.Service
	retryMax int
}

func (s *RetryService) Send(ctx context.Context, tpl string, args []sms.ArgVal, phones ...string) error {
	err := s.svc.Send(ctx, tpl, args, phones...)
	cnt := 1
	for err != nil && cnt < s.retryMax {
		err = s.svc.Send(ctx, tpl, args, phones...)
		if err == nil {
			return nil
		}
		cnt++
	}
	return errors.New("retry all failed")
}
