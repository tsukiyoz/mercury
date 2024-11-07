package ratelimit

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lazywoo/mercury/pkg/ratelimit"
)

type Builder struct {
	prefix  string
	limiter ratelimit.Limiter
}

//go:embed slide_window.lua
var luaScript string

func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	return b.limiter.Limit(ctx, key)
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{
		prefix:  "ip-limiter",
		limiter: limiter,
	}
}
