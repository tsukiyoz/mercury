package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type CaptchaRedisCache struct {
	client redis.Cmdable
}

func (c *CaptchaRedisCache) Set(ctx context.Context, biz string, phone string, captcha string) error {
	ret, err := c.client.Eval(ctx, luaSetCaptcha, []string{c.key(biz, phone)}, captcha).Int()
	if err != nil {
		return err
	}
	switch ret {
	case 0:
		return nil
	case -1:
		return ErrSetCaptchaTooManyTimes
	default:
		return ErrInternal
	}
}

func (c *CaptchaRedisCache) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	ret, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCaptcha).Int()
	if err != nil {
		return false, ErrInternal
	}
	switch ret {
	case 0:
		return true, nil
	case -1:
		// TODO LOG
		return false, ErrCaptchaVerifyTooManyTimes
	case -2:
		return false, nil
	default:
		// TODO LOG
		return false, ErrUnknownForCode
	}
}

func (c *CaptchaRedisCache) key(biz string, phone string) string {
	return fmt.Sprintf("phone_captcha:%s:%s", biz, phone)
}

func NewCaptchaRedisCache(client redis.Cmdable) CaptchaCache {
	return &CaptchaRedisCache{
		client: client,
	}
}
