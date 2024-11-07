package repository

import (
	"context"

	"github.com/lazywoo/mercury/internal/follow/domain"
	"github.com/lazywoo/mercury/internal/follow/repository/cache"
	"github.com/lazywoo/mercury/internal/follow/repository/dao"
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

func (repo *CachedFollowRepository) ActiveFollowRelation(ctx context.Context, r domain.Relation) error {
	err := repo.dao.UpdateStatus(ctx, r.Followee, r.Follower, dao.RelationStatusActive)
	if err != nil {
		return err
	}
	return repo.cache.Follow(ctx, r)
}

func (repo *CachedFollowRepository) InactiveFollowRelation(ctx context.Context, r domain.Relation) error {
	err := repo.dao.UpdateStatus(ctx, r.Followee, r.Follower, dao.RelationStatusInactive)
	if err != nil {
		return err
	}
	return repo.cache.CancelFollow(ctx, r)
}

func (repo *CachedFollowRepository) GetFollowee(ctx context.Context, follower int64, offset, limit int64) ([]domain.Relation, error) {
	list, err := repo.dao.FolloweeRelationList(ctx, follower, offset, limit)
	if err != nil {
		return nil, err
	}
	return repo.genRelationList(list), nil
}

func (repo *CachedFollowRepository) GetFollower(ctx context.Context, followee int64, offset, limit int64) ([]domain.Relation, error) {
	list, err := repo.dao.FollowerRelationList(ctx, followee, offset, limit)
	if err != nil {
		return nil, err
	}
	return repo.genRelationList(list), nil
}

func (repo *CachedFollowRepository) GetRelation(ctx context.Context, r domain.Relation) (domain.Relation, error) {
	res, err := repo.dao.GetRelationDetail(ctx, repo.toEntity(r))
	if err != nil {
		return domain.Relation{}, nil
	}
	return repo.toDomain(res), nil
}

func (repo *CachedFollowRepository) GetStatics(ctx context.Context, uid int64) (domain.Statics, error) {
	res, err := repo.cache.GetStatics(ctx, uid)
	if err == nil {
		return res, err
	}
	res.FolloweeCount, err = repo.dao.CountFollowee(ctx, uid)
	if err != nil {
		return domain.Statics{}, err
	}
	res.FollowerCount, err = repo.dao.CountFollower(ctx, uid)
	if err != nil {
		return domain.Statics{}, err
	}
	err = repo.cache.SetStatics(ctx, uid, res)
	if err != nil {
		repo.l.Error("cache follow statics failed",
			logger.Error(err),
			logger.Int64("uid", uid),
		)
	}
	return res, nil
}

func (repo *CachedFollowRepository) toDomain(r dao.Relation) domain.Relation {
	return domain.Relation{
		Followee: r.Followee,
		Follower: r.Follower,
	}
}

func (repo *CachedFollowRepository) toEntity(r domain.Relation) dao.Relation {
	return dao.Relation{
		Followee: r.Followee,
		Follower: r.Follower,
	}
}

func (repo *CachedFollowRepository) genRelationList(list []dao.Relation) []domain.Relation {
	res := make([]domain.Relation, 0, len(list))
	for _, v := range list {
		res = append(res, repo.toDomain(v))
	}
	return res
}
