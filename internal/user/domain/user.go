package domain

import (
	"time"
)

type User struct {
	Id         int64
	Email      string
	Password   string
	NickName   string
	Phone      string
	AboutMe    string
	WechatInfo WechatInfo
	Birthday   time.Time
	Ctime      time.Time
	Utime      time.Time
}

type WechatInfo struct {
	OpenID  string
	UnionID string
}
