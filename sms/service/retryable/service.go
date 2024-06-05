package retryable

import (
	"context"
	"errors"

	"github.com/lazywoo/mercury/sms/service"
)

type RetryService struct {
	svc      service.Service
	retryMax int
}

func (s *RetryService) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	err := s.svc.Send(ctx, tpl, target, args, values)
	cnt := 1
	for err != nil && cnt < s.retryMax {
		err = s.svc.Send(ctx, tpl, target, args, values)
		if err == nil {
			return nil
		}
		cnt++
	}
	return errors.New("retry all failed")
}
