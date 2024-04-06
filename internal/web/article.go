package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	interactivev1 "github.com/tsukaychan/mercury/api/proto/gen/interactive/v1"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/tsukaychan/mercury/internal/domain"
	"github.com/tsukaychan/mercury/internal/service"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"github.com/tsukaychan/mercury/pkg/ginx"
	"github.com/tsukaychan/mercury/pkg/logger"
	"golang.org/x/sync/errgroup"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	articleSvc service.ArticleService
	intrSvc    interactivev1.InteractiveServiceClient
	logger     logger.Logger

	biz string
}

func NewArticleHandler(articleSvc service.ArticleService, intrSvc interactivev1.InteractiveServiceClient, logger logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		articleSvc: articleSvc,
		intrSvc:    intrSvc,
		logger:     logger,
		biz:        "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)

	// creator
	g.POST("/list", ginx.WrapClaimsAndReq[ListReq, ijwt.UserClaims](h.List))
	g.GET("/detail/:id", ginx.WrapClaims[ijwt.UserClaims](h.Detail))

	pub := server.Group("/pub")
	pub.GET("/:id", ginx.WrapClaims[ijwt.UserClaims](h.PubDetail))
	pub.POST("/like", ginx.WrapClaimsAndReq[LikeReq, ijwt.UserClaims](h.Like))
	pub.POST("/favorite", ginx.WrapClaimsAndReq[FavoriteReq, ijwt.UserClaims](h.Favorite))
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c := ctx.MustGet("user")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		h.logger.Error("no user session found")
	}

	id, err := h.articleSvc.Save(ctx, req.toDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		h.logger.Error("save article failed", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{Data: id})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c := ctx.MustGet("user")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		h.logger.Error("no user session found")
	}

	id, err := h.articleSvc.Publish(ctx, req.toDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		h.logger.Error("publish article failed", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{Data: id})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c := ctx.MustGet("user")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		h.logger.Error("no user session found")
	}

	err := h.articleSvc.Withdraw(ctx, req.Id, claims.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		h.logger.Error("withdraw article failed", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{Msg: "success"})
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (Result, error) {
	if req.Limit > 100 {
		req.Limit = 100
	} else if req.Limit < 0 {
		req.Limit = 0
	}
	atcls, err := h.articleSvc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}
	return Result{
		Data: slice.Map[domain.Article, ArticleVO](atcls, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				// Content: src.Content,
				// Author: src.Author.Name,
				Status: src.Status.ToUint8(),
				Ctime:  src.Ctime.Format(time.DateTime),
				Utime:  src.Utime.Format(time.DateTime),
			}
		}),
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context, uc ijwt.UserClaims) (Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return Result{
			Code: 4,
			Msg:  "invalid params",
		}, err
	}

	atcl, err := h.articleSvc.GetById(ctx, id)
	if err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}

	if atcl.Author.Id != uc.Uid {
		return Result{
			Code: 4,
			Msg:  "invalid params",
		}, fmt.Errorf("illegal access resources, user_id: %d", uc.Uid)
	}

	return Result{
		Data: ArticleVO{
			Id:      atcl.Id,
			Title:   atcl.Title,
			Status:  atcl.Status.ToUint8(),
			Content: atcl.Content,
			Ctime:   atcl.Ctime.Format(time.DateTime),
			Utime:   atcl.Utime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context, uc ijwt.UserClaims) (Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Error("invalid params", logger.Error(err))
		return Result{
			Code: 4,
			Msg:  "invalid params",
		}, fmt.Errorf("get article details %d failed", id)
	}

	var (
		eg       errgroup.Group
		atcl     domain.Article
		intrResp *interactivev1.GetResponse
	)

	eg.Go(func() error {
		var er error
		atcl, er = h.articleSvc.GetPublishedById(ctx, id, uc.Uid)
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
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, fmt.Errorf("get article details failed, error: %w", err)
	}

	if atcl.Author.Id != uc.Uid && atcl.Status == domain.ArticleStatusPrivate {
		return Result{
			Code: 4,
			Msg:  "invalid params",
		}, fmt.Errorf("illegal access resources, user_id: %d", uc.Uid)
	}

	intr := intrResp.Interactive

	return Result{
		Data: ArticleVO{
			Id:          atcl.Id,
			Title:       atcl.Title,
			Content:     atcl.Content,
			Status:      atcl.Status.ToUint8(),
			Author:      atcl.Author.Name,
			LikeCnt:     intr.LikeCnt,
			FavoriteCnt: intr.FavoriteCnt,
			ReadCnt:     intr.ReadCnt,
			Liked:       intr.Liked,
			Favorited:   intr.Favorited,
			Ctime:       atcl.Ctime.Format(time.DateTime),
			Utime:       atcl.Utime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (Result, error) {
	var err error
	if req.Like {
		// err = h.intrSvc.Like(ctx, h.biz, req.Id, uc.Uid)
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
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}

	return Result{Msg: "success"}, err
}

func (h *ArticleHandler) Favorite(ctx *gin.Context, req FavoriteReq, uc ijwt.UserClaims) (Result, error) {
	_, err := h.intrSvc.Favorite(ctx, &interactivev1.FavoriteRequest{
		Biz:   h.biz,
		BizId: req.Id,
		Uid:   uc.Uid,
		Fid:   req.Fid,
	})
	if err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}
	return Result{Msg: "success"}, nil
}
