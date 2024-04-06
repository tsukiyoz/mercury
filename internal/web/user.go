/**
 * @author tsukiyo
 * @date 2023-08-06 12:45
 */

package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tsukaychan/mercury/internal/errs"

	"github.com/tsukaychan/mercury/internal/domain"
	"github.com/tsukaychan/mercury/internal/repository"
	"github.com/tsukaychan/mercury/internal/service"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"github.com/tsukaychan/mercury/pkg/ginx"
	"github.com/tsukaychan/mercury/pkg/logger"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	bizLogin             = "login"
	userIdKey            = "userId"
	emailRegexPattern    = "[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[\\w](?:[\\w-]*[\\w])?"
	passwordRegexPattern = "^(?![a-zA-Z]+$)(?!\\d+$)(?![^\\da-zA-Z\\s]+$).{8,72}$"
)

var _ handler = (*UserHandler)(nil)

type UserHandler struct {
	userService    service.UserService
	captchaService service.CaptchaService
	emailExp       *regexp.Regexp
	passwordExp    *regexp.Regexp
	ijwt.Handler
	logger logger.Logger
}

func NewUserHandler(userService service.UserService, captchaService service.CaptchaService, jwtHandler ijwt.Handler, l logger.Logger) *UserHandler {
	return &UserHandler{
		userService:    userService,
		captchaService: captchaService,
		emailExp:       regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordExp:    regexp.MustCompile(passwordRegexPattern, regexp.None),
		Handler:        jwtHandler,
		logger:         l,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", h.SignUp)
	ug.POST("/login", ginx.WrapReq[LoginReq](h.LoginJWT))
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.ProfileJWT)
	ug.POST("/logout", h.LogoutJWT)
	ug.POST("/login_sms/captcha/send", h.SendLoginCaptcha)
	ug.POST("/login_sms", ginx.WrapReq[LoginSMSReq](h.LoginSMS))
	ug.POST("/refresh_token", h.RefreshToken)
}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	if err := h.ClearToken(ctx); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "logout success",
	})
}

func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	refreshToken := h.ExtractJWTToken(ctx)

	var refreshClaims ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RtKey, nil
	})

	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err = h.CheckSession(ctx, refreshClaims.Ssid); err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if h.SetJWTToken(ctx, refreshClaims.Uid, refreshClaims.Ssid) != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "refresh success",
	})
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}
	if !isEmail {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "email format invalid",
		})
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "passwords doesn't match",
		})
		return
	}

	isPassword, err := h.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}
	if !isPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "password format invalid",
		})
		return
	}

	err = h.userService.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
		Ctime:    time.Now(),
		Utime:    time.Now(),
	})
	if err == service.ErrUserDuplicate {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "the email has been registered",
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
	ctx.JSON(http.StatusOK, Result{Msg: "success"})
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string
		Password string
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := h.userService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "incorrect account or password",
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

	ss := sessions.Default(ctx)
	ss.Set(userIdKey, user.Id)
	ss.Options(sessions.Options{
		Secure:   true,
		HttpOnly: true,
		MaxAge:   30,
	})
	err = ss.Save()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "login success",
	})
}

type LoginReq struct {
	Email    string
	Password string
}

func (h *UserHandler) LoginJWT(ctx *gin.Context, req LoginReq) (Result, error) {
	user, err := h.userService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		return Result{
			Code: 4,
			Msg:  "incorrect account or password",
		}, nil
	}
	if err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, nil
	}

	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, nil
	}

	return Result{
		Msg: "login success",
	}, nil
}

func (h *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "logout success")
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"about_me"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	if req.Nickname == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "nickname cannot be empty",
		})
		return
	}
	if len(req.AboutMe) > 1024 {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "about me too long",
		})
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "incorrect date format",
		})
		return
	}

	claims := ctx.MustGet("user").(*ijwt.UserClaims)
	err = h.userService.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       claims.Uid,
		NickName: req.Nickname,
		AboutMe:  req.AboutMe,
		Birthday: birthday,
		Utime:    time.Now(),
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "success",
	})
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		NickName string
		Birthday string
		AboutMe  string
	}
	uid := sessions.Default(ctx).Get(userIdKey).(int64)
	user, err := h.userService.Profile(ctx, uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: Profile{
			Email:    user.Email,
			Phone:    user.Phone,
			NickName: user.NickName,
			Birthday: user.Birthday.Format(time.DateOnly),
			AboutMe:  user.AboutMe,
		},
	})
}

func (h *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		NickName string
		Birthday string
		AboutMe  string
	}

	claims := ctx.MustGet("user").(*ijwt.UserClaims)

	user, err := h.userService.Profile(ctx, claims.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: Profile{
			Email:    user.Email,
			Phone:    user.Phone,
			NickName: user.NickName,
			Birthday: user.Birthday.Format(time.DateOnly),
			AboutMe:  user.AboutMe,
		},
	})
}

func (h *UserHandler) SendLoginCaptcha(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// TODO reg validate
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "please input your phone number",
		})
		return
	}

	err := h.captchaService.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "send success",
		})
		return
	case repository.ErrCaptchaSendTooManyTimes:
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "send too often, please try again later",
		})
		return
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}
}

type LoginSMSReq struct {
	Phone   string `json:"phone"`
	Captcha string `json:"captcha"`
}

func (h *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (Result, error) {
	ok, err := h.captchaService.Verify(ctx, bizLogin, req.Phone, req.Captcha)
	switch err {
	case nil:
		break
	case repository.ErrCaptchaVerifyTooManyTimes:
		return Result{
			Code: 2,
			Msg:  "verify too many times, please resend the captcha",
		}, nil
	default:
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, fmt.Errorf("user's phone number login failed, error: %w", err)
	}

	if !ok {
		return Result{
			Code: 4,
			Msg:  "captcha invalidated",
		}, nil
	}

	user, err := h.userService.FindOrCreate(ctx, req.Phone)
	if err != nil && err != repository.ErrUserDuplicate {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, fmt.Errorf("user login or register failed, error: %w", err)
	}

	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		return Result{
			Code: 5,
			Msg:  "internal error",
		}, fmt.Errorf("set login token failed, error: %w", err)
	}

	return Result{
		Code: 2,
		Msg:  "captcha validate success",
	}, nil
}
