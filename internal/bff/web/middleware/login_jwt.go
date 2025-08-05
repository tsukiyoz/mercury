package middleware

import (
	"net/http"

	ijwt "github.com/tsukiyo/mercury/internal/bff/web/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
	ignorePaths []string
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: jwtHdl,
	}
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

		signedString := l.ExtractJWTToken(ctx)
		claims := &ijwt.UserClaims{}
		token, err := jwt.ParseWithClaims(signedString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S"), nil
		})
		if err != nil || token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if claims.UserAgent != ctx.Request.UserAgent() {
			// log
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// TODO
		// fallback if redis go wrong

		if err = l.CheckSession(ctx, claims.Ssid); err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", claims)
	}
}
