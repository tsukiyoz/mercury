package captcha

import (
	"context"
	"fmt"
	"github.com/coocood/freecache"
	"strconv"
)

// CaptchaLocalCache development environment, ensure the concurrency safe on single machine
type CaptchaLocalCache struct {
	client *freecache.Cache
}

func (c *CaptchaLocalCache) Set(ctx context.Context, biz string, phone string, code string) error {
	key, cntKey := []byte(c.key(biz, phone)), []byte(c.cntKey(biz, phone))
	ttl, err := c.client.TTL(key)

	if ttl == 0 && err == nil {
		// key exist, but has no expiration
		return ErrInternal
	} else if err != nil && err == freecache.ErrNotFound || ttl < 540 {
		err := c.client.Set(key, []byte(code), 600)
		if err != nil {
			return ErrInternal
		}
		err = c.client.Set(cntKey, []byte("3"), 600)
		if err != nil {
			return ErrInternal
		}
		return nil
	} else {
		return ErrSetCaptchaTooManyTimes
	}
}

func (c *CaptchaLocalCache) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	key, cntKey := []byte(c.key(biz, phone)), []byte(c.cntKey(biz, phone))
	captcha, err := c.client.Get(key)
	if err != nil {
		return false, ErrInternal
	}

	cnt, ttl, err := c.getCntAndTTL(cntKey)
	if err != nil {
		return false, ErrInternal
	}
	if cnt <= 0 {
		return false, ErrCaptchaVerifyTooManyTimes
	}

	if string(captcha) == inputCaptcha {
		err = c.client.Set(cntKey, []byte("-1"), ttl)
		if err != nil {
			return false, ErrInternal
		}
		return true, nil
	} else {
		err = c.client.Set(cntKey, []byte(strconv.Itoa(cnt-1)), ttl)
		if err != nil {
			return false, ErrInternal
		}
		return false, nil
	}
}

func (c *CaptchaLocalCache) key(biz string, phone string, args ...string) string {
	return fmt.Sprintf("phone_captcha:%s:%s", biz, phone)
}

func (c *CaptchaLocalCache) cntKey(biz string, phone string) string {
	return fmt.Sprintf("phone_captcha:%s:%s:cnt", biz, phone)
}

func (c *CaptchaLocalCache) getCntAndTTL(cntKey []byte) (int, int, error) {
	cnt, err := c.client.Get(cntKey)
	if err != nil {
		return 0, 0, err
	}
	ttl, err := c.client.TTL(cntKey)
	if err != nil {
		return 0, 0, err
	}
	cntRet, err := strconv.Atoi(string(cnt))
	if err != nil {
		return 0, 0, err
	}
	return cntRet, int(ttl), nil
}

func NewCaptchaLocalCache(client *freecache.Cache) CaptchaCache {
	return &CaptchaLocalCache{
		client: client,
	}
}
