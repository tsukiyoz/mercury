package repository

import (
	"context"
	"github.com/lazywoo/mercury/follow/domain"
	"github.com/lazywoo/mercury/follow/repository/cache"
	"github.com/lazywoo/mercury/follow/repository/dao"
	"github.com/lazywoo/mercury/pkg/logger"
)

var _ FollowRepository = (*CachedFollowRepository)(nil)

type CachedFollowRepository struct {
	dao   dao.FollowDAO
	cache cache.FollowCache
	l     logger.Logger
}

func NewCachedFollowRepository(
	dao dao.FollowDAO,
	cache cache.FollowCache,
	l logger.Logger,
) FollowRepository {
	return &CachedFollowRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (c CachedFollowRepository) ActiveFollowRelation(ctx context.Context, r domain.Relation) error {
	//TODO implement me
	panic("implement me")
}

func (c CachedFollowRepository) InactiveFollowRelation(ctx context.Context, r domain.Relation) error {
	//TODO implement me
	panic("implement me")
}

func (c CachedFollowRepository) GetFollowee(ctx context.Context, follower int64, offset, limit int64) ([]domain.Relation, error) {
	//TODO implement me
	panic("implement me")
}

func (c CachedFollowRepository) GetFollower(ctx context.Context, followee int64, offset, limit int64) ([]domain.Relation, error) {
	//TODO implement me
	panic("implement me")
}

func (c CachedFollowRepository) GetRelation(ctx context.Context, followee int64, follower int64) (domain.Relation, error) {
	//TODO implement me
	panic("implement me")
}

func (c CachedFollowRepository) GetStatics(ctx context.Context, uid int64) (domain.Statics, error) {
	//TODO implement me
	panic("implement me")
}
