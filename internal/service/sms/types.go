package sms

import "context"

type Service interface {
	Send(ctx context.Context, biz string, args []ArgVal, phones ...string) error
}

type ArgVal struct {
	Val  string
	Name string
}
