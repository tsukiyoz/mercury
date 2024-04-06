package web

import "github.com/tsukaychan/mercury/internal/domain"

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
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}

type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}

type FavoriteReq struct {
	Id  int64 `json:"id"`
	Fid int64 `json:"fid"`
}
