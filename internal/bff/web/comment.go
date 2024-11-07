package web

import (
	"strconv"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	ijwt "github.com/lazywoo/mercury/internal/bff/web/jwt"
	"github.com/lazywoo/mercury/pkg/ginx"

	commentv1 "github.com/lazywoo/mercury/pkg/api/comment/v1"
)

var _ handler = (*CommentHandler)(nil)

type CommentHandler struct {
	commentSvc commentv1.CommentServiceClient
}

func NewCommentHandler(commentSvc commentv1.CommentServiceClient) *CommentHandler {
	return &CommentHandler{
		commentSvc: commentSvc,
	}
}

func (c *CommentHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/comments")
	g.POST("/list", ginx.WrapReqAndClaim[GetCommentListReq](c.GetCommentList))
	g.POST("/delete", ginx.WrapReqAndClaim[DeleteCommentReq](c.DeleteComment))
	g.POST("/create", ginx.WrapReqAndClaim[CreateCommentReq](c.CreateComment))
	g.POST("/reply", ginx.WrapReqAndClaim[GetMoreRepliesRequest](c.GetMoreReplies))
}

func (c *CommentHandler) GetCommentList(ctx *gin.Context, req GetCommentListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := c.commentSvc.GetCommentList(ctx, &commentv1.CommentListRequest{
		Biz:   req.Biz,
		BizId: req.BizId,
		MinId: req.MinId,
		Limit: req.Limit,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}
	return ginx.Result{
		Data: slice.Map[*commentv1.Comment, CommentVO](resp.Comments, func(idx int, src *commentv1.Comment) CommentVO {
			return CommentVO{
				Id:      src.Id,
				Uid:     src.Uid,
				Biz:     src.Biz,
				BizId:   src.BizId,
				Content: src.Content,
				Ctime:   src.Ctime.AsTime().Format(time.DateTime),
				Utime:   src.Utime.AsTime().Format(time.DateTime),
			}
		}),
	}, nil
}

func (c *CommentHandler) DeleteComment(ctx *gin.Context, req DeleteCommentReq, uc ijwt.UserClaims) (ginx.Result, error) {
	gCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("user", strconv.FormatInt(uc.Uid, 10)))
	_, err := c.commentSvc.DeleteComment(gCtx, &commentv1.DeleteCommentRequest{
		Id: req.Id,
	})
	return ginx.Result{}, err
}

func (c *CommentHandler) CreateComment(ctx *gin.Context, req CreateCommentReq, uc ijwt.UserClaims) (ginx.Result, error) {
	gCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("user", strconv.FormatInt(uc.Uid, 10)))
	_, err := c.commentSvc.CreateComment(gCtx, &commentv1.CreateCommentRequest{
		Comment: &commentv1.Comment{
			Uid:     uc.Uid,
			Biz:     req.Biz,
			BizId:   req.BizId,
			Content: req.Content,
			RootComment: &commentv1.Comment{
				Id: req.RootID,
			},
			ParentComment: &commentv1.Comment{
				Id: req.ParentID,
			},
		},
	})
	return ginx.Result{}, err
}

func (c *CommentHandler) GetMoreReplies(ctx *gin.Context, req GetMoreRepliesRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	gCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("user", strconv.FormatInt(uc.Uid, 10)))
	resp, err := c.commentSvc.GetMoreReplies(gCtx, &commentv1.GetMoreRepliesRequest{
		Rid:   req.Rid,
		MaxId: req.MaxID,
		Limit: req.Limit,
	})
	if err != nil {
		return ginx.Result{}, err
	}
	return ginx.Result{
		Data: slice.Map[*commentv1.Comment, CommentVO](resp.Replies, func(idx int, src *commentv1.Comment) CommentVO {
			return CommentVO{
				Id:      src.Id,
				Uid:     src.Uid,
				Biz:     src.Biz,
				BizId:   src.BizId,
				Content: src.Content,
				Ctime:   src.Ctime.AsTime().Format(time.DateTime),
				Utime:   src.Utime.AsTime().Format(time.DateTime),
			}
		}),
	}, nil
}
