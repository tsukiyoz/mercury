package domain

import "github.com/tsukaychan/webook/internal/service/sms"

type AsyncSms struct {
	Id       int64
	TplId    string
	Args     []sms.ArgVal
	Phones   []string
	RetryMax int
}
