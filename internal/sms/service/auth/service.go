package auth

import (
	"context"
	"errors"

	"github.com/tsukiyo/mercury/internal/sms/service"

	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc service.Service
	key string
}

// func (s *SMSService) GenerateToken(ctx context.Context, tplId string) (string, error) {

// }

// Send biz must be a token representing the business side applied offline
func (s *SMSService) Send(ctx context.Context, biz string, target string, args []string, values []string) error {
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
	return s.svc.Send(ctx, tc.Tpl, target, args, values)
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}
