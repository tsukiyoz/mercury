package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	// AtKey access_token key
	AtKey = []byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S")
	// RtKey refresh_token key
	RtKey = []byte("qtzAG6HbKpExQKpsQ9dz7vZGhMnb4g86")
)

var _ Handler = (*RedisJWTHandler)(nil)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd: cmd,
	}
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, userId int64) error {
	ssid := uuid.New().String()
	if err := h.SetJWTToken(ctx, userId, ssid); err != nil {
		return err
	}
	if err := h.setRefreshToken(ctx, userId, ssid); err != nil {
		return err
	}
	return nil
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

func (h *RedisJWTHandler) setRefreshToken(ctx *gin.Context, userId int64, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  userId,
		Ssid: ssid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedString, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-refresh-token", signedString)
	return nil
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, userId int64, ssid string) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
		Uid:       userId,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedString, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-jwt-token", signedString)

	return nil
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	logout, err := h.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if logout > 0 {
		return errors.New("user is logouted")
	}
	return nil
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")

	c, _ := ctx.Get("user")
	claims, ok := c.(*UserClaims)
	if !ok {
		return errors.New("not claims in context")
	}

	return h.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()
}

func (h *RedisJWTHandler) ExtractJWTToken(ctx *gin.Context) string {
	token := ctx.GetHeader("Authorization")

	segs := strings.Split(token, " ")
	if len(segs) != 2 {
		return ""
	}

	return segs[1]
}
