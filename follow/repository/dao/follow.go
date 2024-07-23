package dao

import (
	"context"
	"gorm.io/gorm"
)

var _ FollowDAO = (*GORMFollowDAO)(nil)

type GORMFollowDAO struct {
	db *gorm.DB
}

func NewGORMFollowDAO(db *gorm.DB) FollowDAO {
	return &GORMFollowDAO{
		db: db,
	}
}

func (G *GORMFollowDAO) FolloweeRelationList(ctx context.Context, follower int64, offset, limit int64) ([]Relation, error) {
	//TODO implement me
	panic("implement me")
}

func (G *GORMFollowDAO) FollowerRelationList(ctx context.Context, follower int64, offset, limit int64) ([]Relation, error) {
	//TODO implement me
	panic("implement me")
}

func (G *GORMFollowDAO) GetRelationDetail(ctx context.Context, followee, follower int64) (Relation, error) {
	//TODO implement me
	panic("implement me")
}

func (G *GORMFollowDAO) CreateRelation(ctx context.Context, r Relation) error {
	//TODO implement me
	panic("implement me")
}

func (G *GORMFollowDAO) UpdateStatus(ctx context.Context, followee, follower int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}

func (G *GORMFollowDAO) CountFollowee(ctx context.Context, uid int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (G *GORMFollowDAO) CountFollower(ctx context.Context, uid int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}
