/**
 * @author tsukiyo
 * @date 2023-08-12 13:27
 */

package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

type loginMiddlewareBuilder struct {
	ignorePaths []string
}

func (l *loginMiddlewareBuilder) IgnorePaths(paths ...string) *loginMiddlewareBuilder {
	l.ignorePaths = append(l.ignorePaths, paths...)
	return l
}

func (l *loginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, p := range l.ignorePaths {
			if c.Request.URL.Path == p {
				return
			}
		}
		ss := sessions.Default(c)
		id := ss.Get("user_id")
		if id == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func NewLoginMiddlewareBuilder() *loginMiddlewareBuilder {
	return &loginMiddlewareBuilder{}
}
