package web

type CommentVO struct {
	Id       int64  `json:"id"`
	Uid      int64  `json:"uid"`
	Biz      string `json:"biz"`
	BizId    int64  `json:"biz_id"`
	Content  string `json:"content"`
	RootID   int64  `json:"root_id"`
	ParentID int64  `json:"parent_id"`
	Ctime    string `json:"ctime"`
	Utime    string `json:"utime"`
}

type GetCommentListReq struct {
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
	MinId int64  `json:"min_id"`
	Limit int64  `json:"limit"`
}

type DeleteCommentReq struct {
	Id  int64
	Uid int64
}

type CreateCommentReq struct {
	Id       int64  `json:"id"`
	Uid      int64  `json:"uid"`
	Biz      string `json:"biz"`
	BizId    int64  `json:"biz_id"`
	Content  string `json:"content"`
	RootID   int64  `json:"root_id"`
	ParentID int64  `json:"parent_id"`
	Ctime    string `json:"ctime"`
	Utime    string `json:"utime"`
}

type GetMoreRepliesRequest struct {
	Rid   int64 `json:"rid"`
	MaxID int64 `json:"max_id"`
	Limit int64 `json:"limit"`
}
