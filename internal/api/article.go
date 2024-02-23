package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/pkg/ginx"
	"github.com/tsukaychan/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	articleSvc service.ArticleService
	intrSvc    service.InteractiveService
	logger     logger.Logger

	biz string
}

func NewArticleHandler(articleSvc service.ArticleService, intrSvc service.InteractiveService, logger logger.Logger) *ArticleHandler {
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
	pub.POST("/collect", ginx.WrapClaimsAndReq[CollectReq, ijwt.UserClaims](h.Collect))
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
		eg   errgroup.Group
		atcl domain.Article
		intr domain.Interactive
	)

	eg.Go(func() error {
		var er error
		atcl, er = h.articleSvc.GetPublishedById(ctx, id)
		return er
	})

	eg.Go(func() error {
		var er error
		intr, er = h.intrSvc.Get(ctx, h.biz, id, uc.Uid)
		return er
	})

	err = eg.Wait()
	if err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, fmt.Errorf("get articl details failed, error: %w", err)
	}

	if atcl.Author.Id != uc.Uid && atcl.Status == domain.ArticleStatusPrivate {
		return Result{
			Code: 4,
			Msg:  "invalid params",
		}, fmt.Errorf("illegal access resources, user_id: %d", uc.Uid)
	}

	go func() {
		if gerr := h.intrSvc.IncrReadCnt(ctx, h.biz, atcl.Id); gerr != nil {
			h.logger.Error("increase read cnt failed",
				logger.Int64("article_id", atcl.Id),
				logger.Error(err),
			)
		}
	}()

	return Result{
		Data: ArticleVO{
			Id:         atcl.Id,
			Title:      atcl.Title,
			Content:    atcl.Content,
			Status:     atcl.Status.ToUint8(),
			Author:     atcl.Author.Name,
			LikeCnt:    intr.LikeCnt,
			CollectCnt: intr.CollectCnt,
			ReadCnt:    intr.ReadCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			Ctime:      atcl.Ctime.Format(time.DateTime),
			Utime:      atcl.Utime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (Result, error) {
	var err error
	if req.Like {
		err = h.intrSvc.Like(ctx, h.biz, req.Id, uc.Uid)
	} else {
		err = h.intrSvc.CancelLike(ctx, h.biz, req.Id, uc.Uid)
	}

	if err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}

	return Result{Msg: "success"}, err
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc ijwt.UserClaims) (Result, error) {
	err := h.intrSvc.Collect(ctx, h.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, err
	}
	return Result{Msg: "success"}, nil
}
