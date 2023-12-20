package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/internal/service"
	"webook/internal/service/oauth2/wechat"
)

var _ handler = (*OAuth2WechatHandler)(nil)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	*JWTHandler
}

func NewOAuth2Handler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	url, err := h.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "generate scan login url failed",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	// verify wechat code
	code := ctx.Query("code")
	state := ctx.Query("state")
	info, err := h.svc.VerifyCode(ctx, code, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	user, err := h.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	err = h.setJWTToken(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "success",
	})
}
