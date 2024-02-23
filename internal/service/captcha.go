package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/tsukaychan/webook/internal/repository"
	"github.com/tsukaychan/webook/internal/service/sms"
)

var (
	ErrCodeSendTooManyTimes   = repository.ErrCaptchaSendTooManyTimes
	ErrCodeVerifyTooManyTimes = repository.ErrCaptchaVerifyTooManyTimes
)

var _ CaptchaService = (*captchaService)(nil)

//go:generate mockgen -source=./captcha.go -package=svcmocks -destination=mocks/captcha.mock.go CaptchaService
type CaptchaService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error)
}

type captchaService struct {
	repo    repository.CaptchaRepository
	smsSvc  sms.Service
	tplId   string
	argName string
}

func NewCaptchaService(repo repository.CaptchaRepository, smsSvc sms.Service) CaptchaService {
	return &captchaService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *captchaService) Send(ctx context.Context, biz string, phone string) error {
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

func (svc *captchaService) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCaptcha)
}

func (svc *captchaService) generateCaptcha() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
