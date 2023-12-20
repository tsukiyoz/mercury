package api

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var JWTKey = []byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid          int64
	RefreshCount int64
	UserAgent    string
}

type JWTHandler struct {
}

func (h *JWTHandler) setJWTToken(ctx *gin.Context, userId int64) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
		Uid:          userId,
		RefreshCount: 1,
		UserAgent:    ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedString, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", signedString)
	return nil
}
