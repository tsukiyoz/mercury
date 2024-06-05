package ioc

import (
	"github.com/coocood/freecache"
)

func InitLocalCache() *freecache.Cache {
	client := freecache.NewCache(100 * 1024 * 1024)
	return client
}
