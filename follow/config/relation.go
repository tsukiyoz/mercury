package config

type Relation struct {
	Followee int64 // 被关注者
	Follower int64 // 关注者
}

type Statics struct {
	FollowerCount int64 // 关注者的数量
	FolloweeCount int64 // 关注的数量
}
