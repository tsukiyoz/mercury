package service

import (
	"context"
	"github.com/lazywoo/mercury/follow/domain"
	"github.com/lazywoo/mercury/follow/repository"
)

type FollowService interface {
	Follow(ctx context.Context, followee, follower int64) error
	CancelFollow(ctx context.Context, followee, follower int64) error
	GetFollowee(ctx context.Context, follower int64, offset, limit int64) ([]domain.Relation, error)
	GetFollower(ctx context.Context, followee int64, offset, limit int64) ([]domain.Relation, error)
	GetRelation(ctx context.Context, followee, follower int64) (domain.Relation, error)
	GetStatics(ctx context.Context, uid int64) (domain.Statics, error)
}

var _ FollowService = (*followService)(nil)

type followService struct {
	repo repository.FollowRepository
}

func NewFollowService(repo repository.FollowRepository) FollowService {
	return &followService{
		repo: repo,
	}
}

func (f followService) Follow(ctx context.Context, followee, follower int64) error {
	return f.repo.ActiveFollowRelation(ctx, domain.Relation{
		Followee: followee,
		Follower: follower,
	})
}

func (f followService) CancelFollow(ctx context.Context, followee, follower int64) error {
	return f.repo.InactiveFollowRelation(ctx, domain.Relation{
		Followee: followee,
		Follower: follower,
	})
}

func (f followService) GetFollowee(ctx context.Context, follower int64, offset, limit int64) ([]domain.Relation, error) {
	return f.repo.GetFollowee(ctx, follower, offset, limit)
}

func (f followService) GetFollower(ctx context.Context, followee int64, offset, limit int64) ([]domain.Relation, error) {
	return f.repo.GetFollower(ctx, followee, offset, limit)
}

func (f followService) GetRelation(ctx context.Context, followee, follower int64) (domain.Relation, error) {
	return f.repo.GetRelation(ctx, domain.Relation{
		Followee: followee,
		Follower: follower,
	})
}

func (f followService) GetStatics(ctx context.Context, uid int64) (domain.Statics, error) {
	return f.repo.GetStatics(ctx, uid)
}
