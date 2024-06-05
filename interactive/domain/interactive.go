package domain

type Interactive struct {
	Biz         string `json:"biz"`
	BizId       int64  `json:"biz_id"`
	ReadCnt     int64  `json:"read_cnt"`
	LikeCnt     int64  `json:"like_cnt"`
	FavoriteCnt int64  `json:"favorite_cnt"`
	Liked       bool   `json:"liked"`
	Favorited   bool   `json:"favorited"`
}
