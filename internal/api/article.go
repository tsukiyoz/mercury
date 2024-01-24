package api

import (
	"github.com/gin-gonic/gin"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/service"
	"github.com/tsukaychan/webook/pkg/logger"
	"net/http"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	articleSvc service.ArticleService
	logger     logger.Logger
}

func NewArticleHandler(articleSvc service.ArticleService, logger logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		articleSvc: articleSvc,
		logger:     logger,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		h.logger.Error("no user session found")
	}

	id, err := h.articleSvc.Save(ctx, domain.Article{
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		h.logger.Error("save article failed", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "success",
		Data: id,
	})
}
