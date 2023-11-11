package repository

import (
	"context"
	cache "webook/internal/repository/cache/captcha"
)

var (
	ErrCaptchaSendTooManyTimes   = cache.ErrSetCaptchaTooManyTimes
	ErrCaptchaVerifyTooManyTimes = cache.ErrCaptchaVerifyTooManyTimes
)

var _ CaptchaRepository = (*CaptchaCachedRepository)(nil)

type CaptchaRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error)
}

type CaptchaCachedRepository struct {
	cache cache.CaptchaCache
}

func (repo *CaptchaCachedRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CaptchaCachedRepository) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCaptcha)
}

func NewCaptchaCachedRepository(cache cache.CaptchaCache) CaptchaRepository {
	return &CaptchaCachedRepository{
		cache: cache,
	}
}
