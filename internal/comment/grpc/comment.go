package grpc

import (
	"context"
	"math"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	commentv1 "github.com/lazywoo/mercury/api/gen/comment/v1"
	"github.com/lazywoo/mercury/internal/comment/domain"
	"github.com/lazywoo/mercury/internal/comment/service"
)

type CommentServiceServer struct {
	commentv1.CommentServiceServer
	svc service.CommentService
}

func NewCommentServiceServer(svc service.CommentService) *CommentServiceServer {
	return &CommentServiceServer{
		svc: svc,
	}
}

func (c *CommentServiceServer) Register(srv *grpc.Server) {
	commentv1.RegisterCommentServiceServer(srv, c)
}

func (c *CommentServiceServer) GetCommentList(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	minID := request.GetMinId()
	if minID <= 0 {
		minID = math.MaxInt64
	}
	bizComments, err := c.svc.GetCommentList(ctx, request.GetBiz(), request.GetBizId(), minID, request.GetLimit())
	if err != nil {
		return nil, err
	}
	return &commentv1.CommentListResponse{
		Comments: c.toDTO(bizComments),
	}, nil
}

func (c *CommentServiceServer) DeleteComment(ctx context.Context, request *commentv1.DeleteCommentRequest) (*commentv1.DeleteCommentResponse, error) {
	err := c.svc.DeleteComment(ctx, request.GetId())
	return &commentv1.DeleteCommentResponse{}, err
}

func (c *CommentServiceServer) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	comment := request.GetComment()
	if comment.RootComment == nil && comment.ParentComment != nil || comment.ParentComment == nil && comment.RootComment != nil {
		return &commentv1.CreateCommentResponse{}, status.Error(codes.InvalidArgument, "invalid args")
	}
	err := c.svc.CreateComment(ctx, c.toDomain(comment))
	return &commentv1.CreateCommentResponse{}, err
}

func (c *CommentServiceServer) GetMoreReplies(ctx context.Context, request *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	comments, err := c.svc.GetMoreReplies(ctx, request.GetRid(), request.GetMaxId(), request.GetLimit())
	if err != nil {
		return nil, err
	}
	return &commentv1.GetMoreRepliesResponse{
		Replies: c.toDTO(comments),
	}, nil
}

func (c *CommentServiceServer) toDTO(bizComments []domain.Comment) []*commentv1.Comment {
	dtoComments := make([]*commentv1.Comment, 0, len(bizComments))
	for _, bizComment := range bizComments {
		dtoComment := &commentv1.Comment{
			Id:      bizComment.ID,
			Uid:     bizComment.Commentator.ID,
			Biz:     bizComment.Biz,
			BizId:   bizComment.BizID,
			Content: bizComment.Content,
			Ctime:   timestamppb.New(bizComment.CTime),
			Utime:   timestamppb.New(bizComment.UTime),
		}
		if bizComment.RootComment != nil {
			dtoComment.RootComment = &commentv1.Comment{
				Id: bizComment.RootComment.ID,
			}
		}
		if bizComment.ParentComment != nil {
			dtoComment.ParentComment = &commentv1.Comment{
				Id: bizComment.ParentComment.ID,
			}
		}
		dtoComments = append(dtoComments, dtoComment)
	}

	dtoCommentMap := make(map[int64]*commentv1.Comment, len(dtoComments))
	for _, dtoComment := range dtoComments {
		dtoCommentMap[dtoComment.Id] = dtoComment
	}

	for _, bizComment := range bizComments {
		dtoComment := dtoCommentMap[bizComment.ID]
		if dtoComment.RootComment != nil {
			rootComment, ok := dtoCommentMap[dtoComment.RootComment.Id]
			if ok {
				dtoComment.RootComment = rootComment
			}
		}
		if dtoComment.ParentComment != nil {
			parentComment, ok := dtoCommentMap[dtoComment.ParentComment.Id]
			if ok {
				dtoComment.ParentComment = parentComment
			}
		}
	}
	return dtoComments
}

func (c *CommentServiceServer) toDomain(comment *commentv1.Comment) domain.Comment {
	bizComment := domain.Comment{
		ID: comment.GetId(),
		Commentator: domain.User{
			ID: comment.GetUid(),
		},
		Biz:     comment.GetBiz(),
		BizID:   comment.GetBizId(),
		Content: comment.GetContent(),
	}
	if comment.GetRootComment() != nil {
		bizComment.RootComment = &domain.Comment{
			ID: comment.GetRootComment().GetId(),
		}
	}
	if comment.GetParentComment() != nil {
		bizComment.ParentComment = &domain.Comment{
			ID: comment.GetParentComment().GetId(),
		}
	}
	return bizComment
}
