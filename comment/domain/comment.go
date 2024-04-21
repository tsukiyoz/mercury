package domain

import "time"

type Comment struct {
	ID            int64     `json:"id"`
	Commentator   User      `json:"user"`
	Biz           string    `json:"biz"`
	BizID         int64     `json:"biz_id"`
	Content       string    `json:"content"`
	RootComment   *Comment  `json:"root_comment"`
	ParentComment *Comment  `json:"parent_comment"`
	Children      []Comment `json:"children"`
	CTime         time.Time `json:"ctime"`
	UTime         time.Time `json:"utime"`
}

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
