package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSetCaptchaFrequently    = errors.New("send captcha too frequently")
	ErrInternal                = errors.New("internal error")
	ErrCaptchaVerifyFrequently = errors.New("verify captcha too frequently")
	ErrUnknownForCode          = errors.New("unknown error for code")
)

//go:embed lua/set_code.lua
var luaSetCaptcha string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CaptchaCache struct {
	client redis.Cmdable
}

func (c *CaptchaCache) Set(ctx context.Context, biz string, phone string, code string) error {
	ret, err := c.client.Eval(ctx, luaSetCaptcha, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch ret {
	case 0:
		return nil
	case -1:
		return ErrSetCaptchaFrequently
	default:
		return ErrInternal
	}
}

func (c *CaptchaCache) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	ret, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCaptcha).Int()
	if err != nil {
		return false, err
	}
	switch ret {
	case 0:
		return true, nil
	case -1:
		// TODO LOG
		return false, ErrCaptchaVerifyFrequently
	case -2:
		return false, nil
	default:
		// TODO LOG
		return false, ErrUnknownForCode
	}
}

func (c *CaptchaCache) key(biz string, phone string) string {
	return fmt.Sprintf("phone_captcha:%s:%s", biz, phone)
}

func NewCaptchaCache(client redis.Cmdable) CaptchaCache {
	return CaptchaCache{
		client: client,
	}
}
