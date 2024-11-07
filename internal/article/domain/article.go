package domain

import (
	"time"
)

type ArticleStatus uint8

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

func (atcl Article) Abstract() string {
	cs := []rune(atcl.Content)
	if len(cs) < 100 {
		return atcl.Content
	}
	return string(cs[:100]) + "..."
}

func (status ArticleStatus) ToUint8() uint8 {
	return uint8(status)
}

func (status ArticleStatus) Valid() bool {
	return status.ToUint8() > 0
}

func (status ArticleStatus) NonPublished() bool {
	return status != ArticleStatusPublished
}

func (status ArticleStatus) String() string {
	switch status {
	case ArticleStatusUnpublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	case ArticleStatusPrivate:
		return "private"
	default:
		return "unknown"
	}
}

type Author struct {
	Id   int64
	Name string
}
