/**
 * @author tsukiyo
 * @date 2023-08-11 1:15
 */

package domain

import "time"

type User struct {
	ID         int64
	Email      string
	Password   string
	NickName   string
	Phone      string
	AboutMe    string
	WechatInfo WechatInfo
	Birthday   time.Time
	CreateAt   time.Time
	UpdateAt   time.Time
}
