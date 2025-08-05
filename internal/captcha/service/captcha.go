package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/tsukiyo/mercury/internal/captcha/repository"

	smsv1 "github.com/tsukiyo/mercury/api/gen/sms/v1"
)

var (
	ErrCodeSendTooManyTimes   = repository.ErrCaptchaSendTooManyTimes
	ErrCodeVerifyTooManyTimes = repository.ErrCaptchaVerifyTooManyTimes
)

var _ CaptchaService = (*captchaService)(nil)

//go:generate mockgen -source=./captcha.go -package=svcmocks -destination=mocks/captcha.mock.go CaptchaService
type CaptchaService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, captcha string) (bool, error)
}

type captchaService struct {
	repo   repository.CaptchaRepository
	smsSvc smsv1.SmsServiceClient
	tplId  string
}

func NewCaptchaService(repo repository.CaptchaRepository, smsSvc smsv1.SmsServiceClient) CaptchaService {
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
	_, err = svc.smsSvc.Send(ctx, &smsv1.SendRequest{
		TplId:  svc.tplId,
		Target: phone,
		Args:   []string{"captcha"},
		Values: []string{captcha},
	})
	if err != nil {
		// TODO
		return err
	}
	return nil
}

func (svc *captchaService) Verify(ctx context.Context, biz string, phone string, captcha string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, captcha)
	if errors.Is(err, ErrCodeVerifyTooManyTimes) {
		// TODO alarm here
		return false, nil
	}
	return ok, err
}

func (svc *captchaService) generateCaptcha() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
