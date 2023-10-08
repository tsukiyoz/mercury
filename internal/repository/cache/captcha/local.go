package captcha

import (
	"context"
	"fmt"
	"github.com/coocood/freecache"
	"strconv"
	"sync"
)

type LocalCaptchaCache struct {
	mu     sync.Mutex
	client *freecache.Cache
}

func (c *LocalCaptchaCache) Set(ctx context.Context, biz string, phone string, code string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

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

func (c *LocalCaptchaCache) Verify(ctx context.Context, biz string, phone string, inputCaptcha string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

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

func (c *LocalCaptchaCache) key(biz string, phone string, args ...string) string {
	return fmt.Sprintf("phone_captcha:%s:%s", biz, phone)
}

func (c *LocalCaptchaCache) cntKey(biz string, phone string) string {
	return fmt.Sprintf("phone_captcha:%s:%s:cnt", biz, phone)
}

func (c *LocalCaptchaCache) getCntAndTTL(cntKey []byte) (int, int, error) {
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

func NewLocalCaptchaCache(client *freecache.Cache) CaptchaCache {
	return &LocalCaptchaCache{
		client: client,
	}
}
