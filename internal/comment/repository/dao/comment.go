package dao

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

//go:generate mockgen -source=./comment.go -package=daomocks -destination=mocks/comment.mock.go CommentDAO
type CommentDAO interface {
	Insert(ctx context.Context, u Comment) error
	// FindByBiz return first level comment
	FindByBiz(ctx context.Context, biz string,
		bizId, minID, limit int64) ([]Comment, error)
	// FindCommentList if Comment's id = 0, return first level comment.
	// Otherwise, return the corresponding comment and all its replies
	FindCommentList(ctx context.Context, u Comment) ([]Comment, error)
	FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	Delete(ctx context.Context, u Comment) error
	FindOneByIDs(ctx context.Context, id []int64) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, id int64, limit int64) ([]Comment, error)
}

var _ CommentDAO = (*commentDAO)(nil)

type commentDAO struct {
	db *gorm.DB
}

func NewCommentDAO(db *gorm.DB) CommentDAO {
	return &commentDAO{
		db: db,
	}
}

func (c *commentDAO) Insert(ctx context.Context, u Comment) error {
	return c.db.WithContext(ctx).Create(&u).Error
}

func (c *commentDAO) FindByBiz(ctx context.Context, biz string, bizId, minID, limit int64) ([]Comment, error) {
	var comments []Comment
	err := c.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND id < ? AND pid IS NULL", biz, bizId, minID).
		Limit(int(limit)).
		Find(&comments).Error
	return comments, err
}

func (c *commentDAO) FindCommentList(ctx context.Context, u Comment) ([]Comment, error) {
	var res []Comment
	builder := c.db.WithContext(ctx)
	if u.ID == 0 {
		builder = builder.Where("biz = ?", u.Biz).Where("biz_id = ?", u.BizID).Where("root_id is null")
	} else {
		builder = builder.Where("root_id = ? OR id = ?", u.ID, u.ID)
	}
	err := builder.Find(&res).Error
	return res, err
}

func (c *commentDAO) FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("pid = ?", pid).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (c *commentDAO) Delete(ctx context.Context, u Comment) error {
	return c.db.WithContext(ctx).Delete(&Comment{
		ID: u.ID,
	}).Error
}

func (c *commentDAO) FindOneByIDs(ctx context.Context, id []int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("id IN ?", id).First(&res).Error
	return res, err
}

func (c *commentDAO) FindRepliesByRid(ctx context.Context, rid int64, id int64, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("root_id = ? AND id > ?", rid, id).
		Order("id ASC").
		Limit(int(limit)).
		Find(&res).Error
	return res, err
}

type Comment struct {
	ID            int64         `gorm:"column:id;primaryKey" json:"id"`
	UID           int64         `gorm:"column:uid;index" json:"uid"`
	Biz           string        `gorm:"column:biz;index:biz_type_id" json:"biz"`
	BizID         int64         `gorm:"column:biz_id;index:biz_type_id" json:"biz_id"`
	RootID        sql.NullInt64 `gorm:"column:root_id;index" json:"root_id"` // RootID 0 means root comment
	PID           sql.NullInt64 `gorm:"column:pid;index" json:"pid"`         // PID parent comment ID
	ParentComment *Comment      `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE" json:"parent_comment"`
	Content       string        `gorm:"type:text;column:content" json:"content"`
	Ctime         int64         `gorm:"column:ctime;" json:"ctime"`
	Utime         int64         `gorm:"column:utime;" json:"utime"`
}
