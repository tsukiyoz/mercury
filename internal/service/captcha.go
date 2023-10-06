package service

import (
	"context"
	"fmt"
	"math/rand"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

var (
	ErrCodeSendTooManyTimes   = repository.ErrCaptchaSendTooManyTimes
	ErrCodeVerifyTooManyTimes = repository.ErrCaptchaVerifyTooManyTimes
)

var _ CaptchaService = (*CaptchaServiceV1)(nil)

type CaptchaService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error)
}

type CaptchaServiceV1 struct {
	repo    repository.CaptchaRepository
	smsSvc  sms.Service
	tplId   string
	argName string
}

func (svc *CaptchaServiceV1) Send(ctx context.Context, biz string, phone string) error {
	captcha := svc.generateCaptcha()
	err := svc.repo.Store(ctx, biz, phone, captcha)
	if err != nil {
		return err
	}
	err = svc.smsSvc.Send(ctx, svc.tplId, []sms.ArgVal{
		{
			Name: svc.argName,
			Val:  captcha,
		},
	}, phone)
	if err != nil {
		// TODO
		return err
	}
	return nil
}

func (svc *CaptchaServiceV1) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCaptcha)
}

func (svc *CaptchaServiceV1) generateCaptcha() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func NewCaptchaServiceV1(repo repository.CaptchaRepository, smsSvc sms.Service) CaptchaService {
	return &CaptchaServiceV1{
		repo:   repo,
		smsSvc: smsSvc,
	}
}
