package repository

import (
	"context"
	"github.com/lazywoo/mercury/follow/domain"
)

type FollowRepository interface {
	ActiveFollowRelation(ctx context.Context, r domain.Relation) error
	InactiveFollowRelation(ctx context.Context, r domain.Relation) error
	GetFollowee(ctx context.Context, follower int64, offset, limit int64) ([]domain.Relation, error)
	GetFollower(ctx context.Context, followee int64, offset, limit int64) ([]domain.Relation, error)
	GetRelation(ctx context.Context, followee int64, follower int64) (domain.Relation, error)
	GetStatics(ctx context.Context, uid int64) (domain.Statics, error)
}
