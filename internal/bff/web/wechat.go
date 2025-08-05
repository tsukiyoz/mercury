package web

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	oauth2v1 "github.com/tsukiyo/mercury/api/gen/oauth2/v1"
	userv1 "github.com/tsukiyo/mercury/api/gen/user/v1"

	"github.com/tsukiyo/mercury/pkg/ginx"

	ijwt "github.com/tsukiyo/mercury/internal/bff/web/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
)

var _ handler = (*OAuth2WechatHandler)(nil)

type OAuth2WechatHandler struct {
	svc     oauth2v1.Oauth2ServiceClient
	userSvc userv1.UserServiceClient
	ijwt.Handler
	stateKey []byte
	cfg      WechatHandlerConfig
}

type WechatHandlerConfig struct {
	Secure   bool
	HTTPOnly bool
}

func NewOAuth2Handler(svc oauth2v1.Oauth2ServiceClient, userSvc userv1.UserServiceClient, cfg WechatHandlerConfig, jwtHdl ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:      svc,
		userSvc:  userSvc,
		stateKey: []byte("mzkAG8HhKpRROKpsQ6dX7vZGhNnbRg2S"),
		cfg:      cfg,
		Handler:  jwtHdl,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := uuid.New()
	url, err := h.svc.AuthURL(ctx, &oauth2v1.AuthURLRequest{
		State: state,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	if h.setStateCookie(ctx, state) != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, ijwt.StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("jwt-state", tokenStr, 600, "/oauth2/wechat/callback", "", h.cfg.Secure, h.cfg.HTTPOnly)
	return nil
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	// verify WeChat code and state
	code := ctx.Query("code")
	if err := h.verifyState(ctx); err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "login failed",
		})
		return
	}

	info, err := h.svc.VerifyCode(ctx, &oauth2v1.VerifyCodeRequest{
		Code: code,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	resp, err := h.userSvc.FindOrCreateByWechat(ctx, &userv1.FindOrCreateByWechatRequest{
		Info: &userv1.WechatInfo{
			OpenId:  info.OpenId,
			UnionId: info.UnionId,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	if err = h.SetLoginToken(ctx, resp.GetUser().GetId()); err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{Msg: "OK"})
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")

	cookie, err := ctx.Cookie("jwt-state")
	if err != nil {
		return fmt.Errorf("get state cookie failed, %w", err)
	}

	var stateClaims ijwt.StateClaims
	token, err := jwt.ParseWithClaims(cookie, &stateClaims, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("token is expired, %w", err)
	}

	if stateClaims.State != state {
		return errors.New("invalid state")
	}
	return nil
}
