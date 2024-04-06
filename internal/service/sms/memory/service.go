package memory

import (
	"context"
	"fmt"

	"github.com/tsukaychan/mercury/internal/service/sms"
)

type Service struct{}

func (s *Service) Send(ctx context.Context, tpl string, args []sms.ArgVal, phones ...string) error {
	for _, arg := range args {
		fmt.Printf("captcha: %v\n", arg.Val)
	}
	return nil
}

func NewService() *Service {
	return &Service{}
}
