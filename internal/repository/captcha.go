package repository

import (
	"context"
	"webook/internal/repository/cache"
)

var (
	ErrCaptchaSendTooManyTimes   = cache.ErrSetCaptchaTooManyTimes
	ErrCaptchaVerifyTooManyTimes = cache.ErrCaptchaVerifyTooManyTimes
)

type CaptchaRepository struct {
	cache cache.CaptchaCache
}

func (repo *CaptchaRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CaptchaRepository) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCaptcha)
}

func NewCaptchaRepository(cache cache.CaptchaCache) *CaptchaRepository {
	return &CaptchaRepository{
		cache: cache,
	}
}
