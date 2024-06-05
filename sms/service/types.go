package service

import (
	"context"
)

//go:generate mockgen -source=./types.go -package=smsmocks -destination=mocks/sms.mock.go Service
type Service interface {
	Send(ctx context.Context, tpl string, target string, args []string, values []string) error
}
