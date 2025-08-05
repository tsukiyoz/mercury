package web

import (
	"fmt"
	"net/http"
	"time"

	captchav1 "github.com/tsukiyo/mercury/api/gen/captcha/v1"

	"github.com/tsukiyo/mercury/internal/captcha/repository"

	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/tsukiyo/mercury/api/gen/user/v1"

	"github.com/tsukiyo/mercury/internal/user/errs"
	repository2 "github.com/tsukiyo/mercury/internal/user/repository"

	ijwt "github.com/tsukiyo/mercury/internal/bff/web/jwt"
	"github.com/tsukiyo/mercury/pkg/ginx"
	"github.com/tsukiyo/mercury/pkg/logger"

	regexp "github.com/dlclark/regexp2"
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
	userService    userv1.UserServiceClient
	captchaService captchav1.CaptchaServiceClient
	emailExp       *regexp.Regexp
	passwordExp    *regexp.Regexp
	ijwt.Handler
	logger logger.Logger
}

func NewUserHandler(userService userv1.UserServiceClient, captchaService captchav1.CaptchaServiceClient, jwtHandler ijwt.Handler, l logger.Logger) *UserHandler {
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
	ug.POST("/signup", ginx.WrapReq(h.SignUp))
	ug.POST("/login", ginx.WrapReq[LoginReq](h.LoginJWT))
	ug.POST("/edit", ginx.WrapReqAndClaim[EditReq, ijwt.UserClaims](h.Edit))
	ug.GET("/profile", ginx.WrapReqAndClaim[ProfileReq, ijwt.UserClaims](h.ProfileJWT))
	ug.POST("/logout", ginx.WrapClaims[ijwt.UserClaims](h.LogoutJWT))
	ug.POST("/login_sms/captcha/send", ginx.WrapReq[SendLoginCaptchaReq](h.SendLoginCaptcha))
	ug.POST("/login_sms", ginx.WrapReq[LoginSMSReq](h.LoginSMS))
	ug.POST("/refresh_token", h.RefreshToken)
}

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (h *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
	isEmail, err := h.emailExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, err
	}
	if !isEmail {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "email format invalid",
		}, nil
	}

	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "passwords doesn't match",
		}, nil
	}

	isPassword, err := h.passwordExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, err
	}
	if !isPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "password format invalid",
		}, nil
	}

	_, err = h.userService.SignUp(ctx, &userv1.SignUpRequest{
		User: &userv1.User{
			Email:    req.Email,
			Password: req.Password,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{Msg: "OK"}, nil
}

type LoginReq struct {
	Email    string
	Password string
}

func (h *UserHandler) LoginJWT(ctx *gin.Context, req LoginReq) (ginx.Result, error) {
	resp, err := h.userService.Login(ctx.Request.Context(), &userv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return ginx.Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "invalid input",
		}, err
	}

	err = h.SetLoginToken(ctx, resp.User.Id)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{
		Msg: "OK",
	}, nil
}

type EditReq struct {
	Nickname string `json:"nickname"`
	Birthday string `json:"birthday"`
	AboutMe  string `json:"about_me"`
}

func (h *UserHandler) Edit(ctx *gin.Context, req EditReq, uc ijwt.UserClaims) (ginx.Result, error) {
	if req.Nickname == "" {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "nickname cannot be empty",
		}, nil
	}
	if len(req.AboutMe) > 1024 {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "about me too long",
		}, nil
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "incorrect date format",
		}, err
	}

	_, err = h.userService.UpdateNonSensitiveInfo(ctx, &userv1.UpdateNonSensitiveInfoRequest{
		User: &userv1.User{
			Id:       uc.Uid,
			NickName: req.Nickname,
			AboutMe:  req.AboutMe,
			Birthday: timestamppb.New(birthday),
		},
	})
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{
		Msg: "OK",
	}, nil
}

type ProfileReq struct {
	Email    string
	Phone    string
	NickName string
	Birthday string
	AboutMe  string
}

func (h *UserHandler) ProfileJWT(ctx *gin.Context, req ProfileReq, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.userService.Profile(ctx, &userv1.ProfileRequest{Id: uc.Uid})
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{
		Data: ProfileReq{
			Email:    resp.GetUser().GetEmail(),
			Phone:    resp.GetUser().GetPhone(),
			NickName: resp.GetUser().GetNickName(),
			Birthday: resp.GetUser().GetBirthday().AsTime().Format(time.DateOnly),
			AboutMe:  resp.GetUser().GetAboutMe(),
		},
	}, nil
}

func (h *UserHandler) LogoutJWT(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	if err := h.ClearToken(ctx, uc.Ssid); err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, err
	}

	return ginx.Result{
		Msg: "OK",
	}, nil
}

type SendLoginCaptchaReq struct {
	Phone string `json:"phone"`
}

func (h *UserHandler) SendLoginCaptcha(ctx *gin.Context, req SendLoginCaptchaReq) (ginx.Result, error) {
	// TODO reg validate
	if req.Phone == "" {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "invalid input",
		}, nil
	}

	_, err := h.captchaService.Send(ctx, &captchav1.SendRequest{Biz: bizLogin, Phone: req.Phone})
	switch err {
	case nil:
		return ginx.Result{
			Msg: "OK",
		}, nil
	case repository.ErrCaptchaSendTooManyTimes:
		return ginx.Result{
			Msg: "send too often, please try again later",
		}, nil
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, nil
	}
}

type LoginSMSReq struct {
	Phone   string `json:"phone"`
	Captcha string `json:"captcha"`
}

func (h *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (ginx.Result, error) {
	verifyResp, err := h.captchaService.Verify(ctx, &captchav1.VerifyRequest{
		Biz:     bizLogin,
		Phone:   req.Phone,
		Captcha: req.Captcha,
	})
	switch err {
	case nil:
		break
	case repository.ErrCaptchaVerifyTooManyTimes:
		return ginx.Result{
			Code: 4,
			Msg:  "verify too many times, please resend the captcha",
		}, nil
	default:
		return ginx.Result{
			Code: 4,
			Msg:  "internal error",
		}, fmt.Errorf("user's phone number login failed, error: %w", err)
	}

	if !verifyResp.Answer {
		return ginx.Result{
			Code: 4,
			Msg:  "captcha invalidated",
		}, nil
	}

	resp, err := h.userService.FindOrCreate(ctx, &userv1.FindOrCreateRequest{
		Phone: req.Phone,
	})
	if err != nil && err != repository2.ErrUserDuplicate {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, fmt.Errorf("user login or register failed, error: %w", err)
	}

	if err = h.SetLoginToken(ctx, resp.GetUser().GetId()); err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "internal error",
		}, fmt.Errorf("set login token failed, error: %w", err)
	}

	return ginx.Result{
		Msg: "OK",
	}, nil
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

	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}
