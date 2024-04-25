package repository

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tsukaychan/mercury/comment/domain"
	"github.com/tsukaychan/mercury/comment/repository/dao"
	"github.com/tsukaychan/mercury/pkg/logger"
)

type CommentRepository interface {
	FindByBiz(ctx context.Context, biz string,
		bizId, minID, limit int64) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, comment domain.Comment) error
	CreateComment(ctx context.Context, comment domain.Comment) error
	GetCommentByIds(ctx context.Context, id []int64) ([]domain.Comment, error)
	GetMoreReplies(ctx context.Context, rid int64, id int64, limit int64) ([]domain.Comment, error)
}

var _ CommentRepository = (*commentRepository)(nil)

type commentRepository struct {
	dao dao.CommentDAO
	l   logger.Logger
}

func NewCommentRepository(dao dao.CommentDAO, l logger.Logger) CommentRepository {
	return &commentRepository{
		dao: dao,
		l:   l,
	}
}

func (c *commentRepository) FindByBiz(ctx context.Context, biz string, bizId, minID, limit int64) ([]domain.Comment, error) {
	dbComments, err := c.dao.FindByBiz(ctx, biz, bizId, minID, limit)
	if err != nil {
		return nil, err
	}
	bizComments := make([]domain.Comment, 0, len(dbComments))
	var eg errgroup.Group
	downgraded := ctx.Value("downgraded") == "true"
	for _, dbComment := range dbComments {
		dbComment := dbComment
		bizComment := c.toDomain(dbComment)
		bizComments = append(bizComments, bizComment)
		if downgraded {
			continue
		}
		eg.Go(func() error {
			// show only three
			bizComment.Children = make([]domain.Comment, 0, 3)
			rs, err := c.dao.FindRepliesByPid(ctx, dbComment.ID, 0, 3)
			if err != nil {
				c.l.Error("get child comment failed", logger.Error(err))
				return nil
			}
			for _, r := range rs {
				bizComment.Children = append(bizComment.Children, c.toDomain(r))
			}
			return nil
		})
	}
	return bizComments, eg.Wait()
}

func (c *commentRepository) DeleteComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Delete(ctx, c.toEntity(comment))
}

func (c *commentRepository) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Insert(ctx, c.toEntity(comment))
}

func (c *commentRepository) GetCommentByIds(ctx context.Context, id []int64) ([]domain.Comment, error) {
	dbComments, err := c.dao.FindOneByIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	comments := make([]domain.Comment, 0, len(dbComments))
	for _, v := range dbComments {
		comment := c.toDomain(v)
		comments = append(comments, comment)
	}
	return comments, nil
}

func (c *commentRepository) GetMoreReplies(ctx context.Context, rid int64, id int64, limit int64) ([]domain.Comment, error) {
	comments, err := c.dao.FindRepliesByRid(ctx, rid, id, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(comments))
	for _, comment := range comments {
		res = append(res, c.toDomain(comment))
	}
	return res, nil
}

func (c *commentRepository) toDomain(dbComment dao.Comment) domain.Comment {
	bizComment := domain.Comment{
		ID: dbComment.ID,
		Commentator: domain.User{
			ID: dbComment.UID,
		},
		Biz:     dbComment.Biz,
		BizID:   dbComment.BizID,
		Content: dbComment.Content,
		CTime:   time.UnixMilli(dbComment.Ctime),
		UTime:   time.UnixMilli(dbComment.Utime),
	}
	if dbComment.RootID.Valid {
		bizComment.RootComment = &domain.Comment{
			ID: dbComment.RootID.Int64,
		}
	}
	if dbComment.PID.Valid {
		bizComment.ParentComment = &domain.Comment{
			ID: dbComment.PID.Int64,
		}
	}
	return bizComment
}

func (c *commentRepository) toEntity(bizComment domain.Comment) dao.Comment {
	dbComment := dao.Comment{
		ID:      bizComment.ID,
		UID:     bizComment.Commentator.ID,
		Biz:     bizComment.Biz,
		BizID:   bizComment.BizID,
		Content: bizComment.Content,
	}
	if bizComment.RootComment != nil && bizComment.RootComment.ID != 0 {
		dbComment.RootID = sql.NullInt64{
			Int64: bizComment.RootComment.ID,
			Valid: true,
		}
	}
	if bizComment.ParentComment != nil && bizComment.ParentComment.ID != 0 {
		dbComment.PID = sql.NullInt64{
			Int64: bizComment.ParentComment.ID,
			Valid: true,
		}
	}
	dbComment.Ctime = time.Now().UnixMilli()
	dbComment.Utime = time.Now().UnixMilli()
	return dbComment
}
