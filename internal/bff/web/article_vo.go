package web

import (
	articlev1 "github.com/tsukiyo/mercury/api/gen/article/v1"
)

type ArticleVO struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Content  string `json:"content"`
	Status   uint8  `json:"status"`
	Author   string `json:"author"`

	LikeCnt     int64 `json:"like_cnt"`
	FavoriteCnt int64 `json:"favorite_cnt"`
	ReadCnt     int64 `json:"read_cnt"`

	Liked     bool `json:"liked"`
	Favorited bool `json:"favorited"`

	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}

type ListReq struct {
	Offset int32 `json:"offset"`
	Limit  int32 `json:"limit"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDTO(uid int64) *articlev1.Article {
	return &articlev1.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: &articlev1.Author{
			Id: uid,
		},
	}
}

type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}

type FavoriteReq struct {
	Id       int64 `json:"id"`
	Fid      int64 `json:"fid"`
	Favorite bool  `json:"favorite"`
}
