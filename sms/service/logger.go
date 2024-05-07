package service

import (
	"context"
	"strings"

	"github.com/tsukaychan/mercury/pkg/logger"

	"go.uber.org/zap"
)

type LoggerService struct {
	svc Service
	l   logger.Logger
}

func NewService(svc Service, l logger.Logger) Service {
	return &LoggerService{svc: svc, l: l}
}

func (s *LoggerService) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	s.l.Info("send captcha",
		logger.String("biz", tpl),
		logger.String("target", target),
		logger.String("args", strings.Join(args, ",")),
		logger.String("values", strings.Join(values, ",")),
	)
	err := s.svc.Send(ctx, tpl, target, args, values)
	if err != nil {
		zap.L().Debug("send captcha failed", zap.Error(err))
	}
	return err
}
