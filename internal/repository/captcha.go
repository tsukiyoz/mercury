package repository

import (
	"context"
	"webook/internal/repository/cache/captcha"
)

var (
	ErrCaptchaSendTooManyTimes   = captcha.ErrSetCaptchaTooManyTimes
	ErrCaptchaVerifyTooManyTimes = captcha.ErrCaptchaVerifyTooManyTimes
)

var _ CaptchaRepository = (*CachedCaptchaRepository)(nil)

type CaptchaRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error)
}

type CachedCaptchaRepository struct {
	cache captcha.CaptchaCache
}

func (repo *CachedCaptchaRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CachedCaptchaRepository) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCaptcha)
}

func NewCachedCaptchaRepository(cache captcha.CaptchaCache) CaptchaRepository {
	return &CachedCaptchaRepository{
		cache: cache,
	}
}
