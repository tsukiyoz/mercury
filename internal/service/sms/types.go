package sms

import "context"

//go:generate mockgen -source=./types.go -package=smsmocks -destination=mocks/sms.mock.go Service
type Service interface {
	Send(ctx context.Context, biz string, args []ArgVal, phones ...string) error
}

type ArgVal struct {
	Name string
	Val  string
}
