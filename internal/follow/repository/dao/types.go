package dao

import "context"

type FollowDAO interface {
	FolloweeRelationList(ctx context.Context, follower int64, offset, limit int64) ([]Relation, error)
	FollowerRelationList(ctx context.Context, follower int64, offset, limit int64) ([]Relation, error)
	GetRelationDetail(ctx context.Context, r Relation) (Relation, error)
	CreateRelation(ctx context.Context, r Relation) error
	UpdateStatus(ctx context.Context, followee, follower int64, status uint8) error
	CountFollowee(ctx context.Context, uid int64) (int64, error)
	CountFollower(ctx context.Context, uid int64) (int64, error)
}
