package jwt

import "github.com/gin-gonic/gin"

type Handler interface {
	SetLoginToken(ctx *gin.Context, userId int64) error
	SetJWTToken(ctx *gin.Context, userId int64, ssid string) error
	// setRefreshToken(ctx *gin.Context, userId int64, ssid string) error
	CheckSession(ctx *gin.Context, ssid string) error
	ClearToken(ctx *gin.Context) error
	ExtractJWTToken(ctx *gin.Context) string
}
