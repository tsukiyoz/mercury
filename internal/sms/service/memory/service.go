package memory

import (
	"context"
	"fmt"

	"github.com/lazywoo/mercury/internal/sms/service"
)

type Service struct{}

func (s *Service) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	fmt.Printf("send captcha to target[%v]\n", target)
	return nil
}

func NewService() service.Service {
	return &Service{}
}
