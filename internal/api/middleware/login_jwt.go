/**
 * @author tsukiyo
 * @date 2023-08-29 02:57
 */

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
	"webook/internal/api"
)

type LoginJWTMiddlewareBuilder struct {
	ignorePaths []string
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(paths ...string) *LoginJWTMiddlewareBuilder {
	l.ignorePaths = append(l.ignorePaths, paths...)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, p := range l.ignorePaths {
			if ctx.Request.URL.Path == p {
				return
			}
		}
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(token, " ")
		if len(segs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token = segs[1]
		claims := &api.UserClaims{}
		signedString, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S"), nil
		})
		if err != nil || signedString == nil || !signedString.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if claims.UserAgent != ctx.Request.UserAgent() {
			// log
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		currentTime := time.Now()
		if claims.ExpiresAt.Sub(currentTime) < 3*time.Minute*time.Duration(claims.RefreshCount) {
			if claims.RefreshCount < 3600 {
				claims.RefreshCount++
			}
			claims.ExpiresAt = jwt.NewNumericDate(currentTime.Add(time.Minute * 4 * time.Duration(claims.RefreshCount)))
			token, err = signedString.SignedString([]byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S"))
			if err != nil {
				// log
				log.Println("jwt renewal failed")
			}
			ctx.Header("x-jwt-token", token)
		}

		ctx.Set("claims", claims)
	}
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}
