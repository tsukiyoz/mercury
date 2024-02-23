/**
 * @author tsukiyo
 * @date 2023-08-12 13:27
 */

package middleware

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	ignorePaths []string
}

func (l *LoginMiddlewareBuilder) IgnorePaths(paths ...string) *LoginMiddlewareBuilder {
	l.ignorePaths = append(l.ignorePaths, paths...)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(c *gin.Context) {
		for _, p := range l.ignorePaths {
			if c.Request.URL.Path == p {
				return
			}
		}
		sess := sessions.Default(c)
		id := sess.Get("user_id")

		if id == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		sess.Set("user_id", id)
		sess.Options(sessions.Options{
			MaxAge: 30,
		})
		updateTimeVal := sess.Get("update_time")
		currentTime := time.Now()

		if updateTimeVal == nil {
			sess.Set("update_time", currentTime)
			sess.Save()
		}

		updateTime, _ := updateTimeVal.(time.Time)
		if currentTime.Sub(updateTime) > time.Second*15 {
			sess.Set("update_time", currentTime)
			sess.Save()
			fmt.Println("login status refreshed")
		}
	}
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}
