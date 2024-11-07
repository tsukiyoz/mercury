package cache

import (
	"context"

	"github.com/lazywoo/mercury/internal/follow/domain"
)

type FollowCache interface {
	GetStatics(ctx context.Context, uid int64) (domain.Statics, error)
	SetStatics(ctx context.Context, uid int64, s domain.Statics) error
	Follow(ctx context.Context, r domain.Relation) error
	CancelFollow(ctx context.Context, r domain.Relation) error
}
