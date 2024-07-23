package dao

type Relation struct {
	ID       int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Followee int64 `gorm:"type:int(11);not null;uniqueIndex:followee_follower"`
	Follower int64 `gorm:"type:int(11);not null;uniqueIndex:followee_follower"`
	Status   uint8
	Ctime    int64
	Utime    int64
}

const (
	RelationStatusUnknown = iota
	RelationStatusActive
	RelationStatusInactive
)

type Statics struct {
	ID            int64 `gorm:"primaryKey,autoIncrement,colum:id"`
	UID           int64 `gorm:"unique"`
	FolloweeCount int64
	FollowerCount int64
	Ctime         int64
	Utime         int64
}
