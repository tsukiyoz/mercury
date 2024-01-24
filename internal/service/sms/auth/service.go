package auth

import (
	"context"
	"errors"
	"github.com/tsukaychan/webook/internal/service/sms"

	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc sms.Service
	key string
}

// func (s *SMSService) GenerateToken(ctx context.Context, tplId string) (string, error) {

// }

// Send biz must be a token representing the business side applied offline
func (s *SMSService) Send(ctx context.Context, biz string, args []sms.ArgVal, phones ...string) error {
	var tc Claims
	token, err := jwt.ParseWithClaims(biz, &tc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token invalid")
	}
	return s.svc.Send(ctx, tc.Tpl, args, phones...)
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}
