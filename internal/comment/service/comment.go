package service

import (
	"context"

	"github.com/tsukiyo/mercury/internal/comment/domain"
	"github.com/tsukiyo/mercury/internal/comment/repository"
)

type CommentService interface {
	GetCommentList(ctx context.Context, biz string, bizId, minID, limit int64) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	CreateComment(ctx context.Context, comment domain.Comment) error
	GetMoreReplies(ctx context.Context, rid int64, maxID int64, limit int64) ([]domain.Comment, error)
}

var _ CommentService = (*commentService)(nil)

type commentService struct {
	repo repository.CommentRepository
}

func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentService{
		repo: repo,
	}
}

func (c *commentService) GetCommentList(ctx context.Context, biz string, bizId, minID, limit int64) ([]domain.Comment, error) {
	list, err := c.repo.FindByBiz(ctx, biz, bizId, minID, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *commentService) DeleteComment(ctx context.Context, id int64) error {
	return c.repo.DeleteComment(ctx, domain.Comment{
		ID: id,
	})
}

func (c *commentService) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.repo.CreateComment(ctx, comment)
}

func (c *commentService) GetCommentByIds(ctx context.Context, id []int64) ([]domain.Comment, error) {
	return c.repo.GetCommentByIds(ctx, id)
}

func (c *commentService) GetMoreReplies(ctx context.Context, rid int64, maxID int64, limit int64) ([]domain.Comment, error) {
	return c.repo.GetMoreReplies(ctx, rid, maxID, limit)
}
