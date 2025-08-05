package web

import (
	"fmt"
	"strconv"
	"time"

	articlev1 "github.com/tsukiyo/mercury/api/gen/article/v1"

	interactivev1 "github.com/tsukiyo/mercury/api/gen/interactive/v1"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	ijwt "github.com/tsukiyo/mercury/internal/bff/web/jwt"
	"github.com/tsukiyo/mercury/pkg/ginx"
	"github.com/tsukiyo/mercury/pkg/logger"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	articleSvc articlev1.ArticleServiceClient
	intrSvc    interactivev1.InteractiveServiceClient
	l          logger.Logger

	biz string
}

func NewArticleHandler(articleSvc articlev1.ArticleServiceClient, intrSvc interactivev1.InteractiveServiceClient, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		articleSvc: articleSvc,
		intrSvc:    intrSvc,
		l:          l,
		biz:        "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", ginx.WrapReqAndClaim[ArticleReq, ijwt.UserClaims](h.Edit))
	g.POST("/publish", ginx.WrapReqAndClaim[ArticleReq, ijwt.UserClaims](h.Publish))
	g.POST("/withdraw", ginx.WrapReqAndClaim[WithdrawReq, ijwt.UserClaims](h.Withdraw))

	// creator
	g.POST("/list", ginx.WrapReqAndClaim[ListReq, ijwt.UserClaims](h.List))
	g.GET("/detail/:id", ginx.WrapClaims[ijwt.UserClaims](h.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", ginx.WrapClaims[ijwt.UserClaims](h.PubDetail))
	pub.POST("/like", ginx.WrapReqAndClaim[LikeReq, ijwt.UserClaims](h.Like))
	pub.POST("/favorite", ginx.WrapReqAndClaim[FavoriteReq, ijwt.UserClaims](h.Favorite))
}

func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.articleSvc.Save(ctx, &articlev1.SaveRequest{Article: req.toDTO(uc.Uid)})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{
		Data: resp.GetId(),
	}, nil
}

func (h *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.articleSvc.Publish(ctx, &articlev1.PublishRequest{
		Article: req.toDTO(uc.Uid),
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{
		Data: resp.GetId(),
	}, nil
}

type WithdrawReq struct {
	Id int64
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req WithdrawReq, uc ijwt.UserClaims) (ginx.Result, error) {
	_, err := h.articleSvc.Withdraw(ctx, &articlev1.WithdrawRequest{
		Uid: uc.Uid,
		Id:  req.Id,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	if req.Limit > 100 {
		req.Limit = 100
	} else if req.Limit < 0 {
		req.Limit = 0
	}
	listResp, err := h.articleSvc.List(ctx, &articlev1.ListRequest{
		Author: uc.Uid,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}
	return ginx.Result{
		Data: slice.Map[*articlev1.Article, ArticleVO](listResp.Articles, func(idx int, src *articlev1.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract,
				// Content: src.Content,
				// Author: src.Author.Name,
				Status: uint8(src.Status),
				Ctime:  src.Ctime.AsTime().Format(time.DateTime),
				Utime:  src.Utime.AsTime().Format(time.DateTime),
			}
		}),
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "invalid params",
		}, err
	}

	resp, err := h.articleSvc.GetById(ctx, &articlev1.GetByIdRequest{Id: id})
	atcl := resp.GetArticle()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}

	if atcl.GetAuthor().GetId() != uc.Uid {
		return ginx.Result{
			Code: 4,
			Msg:  "invalid params",
		}, fmt.Errorf("illegal access resources, user_id: %d", uc.Uid)
	}

	return ginx.Result{
		Data: ArticleVO{
			Id:      atcl.Id,
			Title:   atcl.Title,
			Status:  uint8(atcl.Status),
			Content: atcl.Content,
			Ctime:   atcl.Ctime.AsTime().Format(time.DateTime),
			Utime:   atcl.Utime.AsTime().Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.l.Error("invalid params", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "invalid params",
		}, fmt.Errorf("get article details %d failed", id)
	}

	var (
		eg       errgroup.Group
		atcl     *articlev1.Article
		intrResp *interactivev1.GetResponse
	)

	eg.Go(func() error {
		var er error
		resp, er := h.articleSvc.GetPublishedById(ctx, &articlev1.GetPublishedByIdRequest{
			Id:  id,
			Uid: uc.Uid,
		})
		atcl = resp.GetArticle()
		return er
	})

	eg.Go(func() error {
		var er error
		intrResp, er = h.intrSvc.Get(ctx, &interactivev1.GetRequest{
			Biz:   h.biz,
			BizId: id,
			Uid:   uc.Uid,
		})
		return er
	})

	err = eg.Wait()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, fmt.Errorf("get article details failed, error: %w", err)
	}

	intr := intrResp.Interactive

	return ginx.Result{
		Data: ArticleVO{
			Id:          atcl.Id,
			Title:       atcl.Title,
			Content:     atcl.Content,
			Status:      uint8(atcl.Status),
			Author:      atcl.Author.Name,
			LikeCnt:     intr.LikeCnt,
			FavoriteCnt: intr.FavoriteCnt,
			ReadCnt:     intr.ReadCnt,
			Liked:       intr.Liked,
			Favorited:   intr.Favorited,
			Ctime:       atcl.Ctime.AsTime().Format(time.DateTime),
			Utime:       atcl.Utime.AsTime().Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		_, err = h.intrSvc.Like(ctx, &interactivev1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
		})
	} else {
		_, err = h.intrSvc.CancelLike(ctx, &interactivev1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
		})
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{Msg: "OK"}, err
}

func (h *ArticleHandler) Favorite(ctx *gin.Context, req FavoriteReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Favorite {
		_, err = h.intrSvc.Favorite(ctx, &interactivev1.FavoriteRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
			Fid:   req.Fid,
		})
	} else {
		_, err = h.intrSvc.CancelFavorite(ctx, &interactivev1.CancelFavoriteRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
			Fid:   req.Fid,
		})
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}
