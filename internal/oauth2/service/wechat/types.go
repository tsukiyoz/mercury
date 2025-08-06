package wechat

import (
	"context"

	"github.com/tsukiyo/mercury/internal/oauth2/domain"
)

//go:generate mockgen -source=./types.go -package=wechatmocks -destination=mocks/svc.mock.go Service
type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}
