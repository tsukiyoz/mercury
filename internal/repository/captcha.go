package repository

import (
	"context"
	cache "github.com/tsukaychan/webook/internal/repository/cache/captcha"
)

var (
	ErrCaptchaSendTooManyTimes   = cache.ErrSetCaptchaTooManyTimes
	ErrCaptchaVerifyTooManyTimes = cache.ErrCaptchaVerifyTooManyTimes
)

var _ CaptchaRepository = (*CachedCaptchaRepository)(nil)

//go:generate mockgen -source=./captcha.go -package=repomocks -destination=mocks/captcha.mock.go CaptchaRepository
type CaptchaRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error)
}

type CachedCaptchaRepository struct {
	cache cache.CaptchaCache
}

func NewCachedCaptchaRepository(cache cache.CaptchaCache) CaptchaRepository {
	return &CachedCaptchaRepository{
		cache: cache,
	}
}

func (repo *CachedCaptchaRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CachedCaptchaRepository) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCaptcha)
}
