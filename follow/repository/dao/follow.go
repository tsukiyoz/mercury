package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
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

func (dao *GORMFollowDAO) FolloweeRelationList(ctx context.Context, follower int64, offset, limit int64) ([]Relation, error) {
	var res []Relation
	err := dao.db.WithContext(ctx).Model(&Relation{}).
		Where("follower = ? AND status = ?", follower, RelationStatusActive).
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&res).Error
	return res, err
}

func (dao *GORMFollowDAO) FollowerRelationList(ctx context.Context, follower int64, offset, limit int64) ([]Relation, error) {
	var res []Relation
	err := dao.db.WithContext(ctx).Model(&Relation{}).
		Where("followee = ? AND status = ?", follower, RelationStatusActive).
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&res).Error
	return res, err
}

func (dao *GORMFollowDAO) GetRelationDetail(ctx context.Context, r Relation) (Relation, error) {
	var res Relation
	err := dao.db.WithContext(ctx).Model(&Relation{}).
		Where("followee = ? AND follower = ? AND status = ?", r.Followee, r.Follower, RelationStatusActive).
		First(&res).Error
	return res, err
}

func (dao *GORMFollowDAO) CreateRelation(ctx context.Context, r Relation) error {
	return dao.db.WithContext(ctx).Model(&Relation{}).Create(r).Error
}

func (dao *GORMFollowDAO) UpdateStatus(ctx context.Context, followee, follower int64, status uint8) error {
	now := time.Now()
	return dao.db.WithContext(ctx).Model(&Relation{}).
		Where("followee = ? AND follower = ?", followee, follower).
		Updates(map[string]any{
			"status": status,
			"utime":  now,
		}).Error
}

func (dao *GORMFollowDAO) CountFollowee(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := dao.db.WithContext(ctx).
		Select("count(followee)").
		Where("follower = ? AND status = ?", uid, RelationStatusActive).
		Count(&res).Error
	return res, err
}

func (dao *GORMFollowDAO) CountFollower(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := dao.db.WithContext(ctx).
		Select("count(follower)").
		Where("followee = ? AND status = ?", uid, RelationStatusActive).
		Count(&res).Error
	return res, err
}
