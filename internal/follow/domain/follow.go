package domain

type Relation struct {
	Followee int64 // 被关注者
	Follower int64 // 关注者
}

type Statics struct {
	FolloweeCount int64 // 关注数量
	FollowerCount int64 // 被关注数量
}
