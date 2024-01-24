/**
 * @author tsukiyo
 * @date 2023-08-06 12:45
 */

package api

import (
	"fmt"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository"
	"github.com/tsukaychan/webook/internal/service"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
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
	cmd redis.Cmdable
}

func NewUserHandler(userService service.UserService, captchaService service.CaptchaService, jwtHandler ijwt.Handler) *UserHandler {
	return &UserHandler{
		userService:    userService,
		captchaService: captchaService,
		emailExp:       regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordExp:    regexp.MustCompile(passwordRegexPattern, regexp.None),
		Handler:        jwtHandler,
	}
}
func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/logout", u.LogoutJWT)
	ug.POST("/login_sms/captcha/send", u.SendLoginCaptcha)
	ug.POST("/login_sms/captcha/validate", u.LoginSMS)
	ug.POST("/refresh_token", u.RefreshToken)
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	if err := u.ClearToken(ctx); err != nil {
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

func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	refreshToken := u.ExtractJWTToken(ctx)

	var refreshClaims ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RtKey, nil
	})

	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err = u.CheckSession(ctx, refreshClaims.Ssid); err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if count, err := u.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", refreshClaims.Ssid)).Result(); err != nil || count > 0 {
		// redis wrong or token is expired
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if u.SetJWTToken(ctx, refreshClaims.Uid, refreshClaims.Ssid) != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "refresh success",
	})
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := u.emailExp.MatchString(req.Email)
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

	isPassword, err := u.passwordExp.MatchString(req.Password)
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

	err = u.userService.SignUp(ctx, domain.User{
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
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "success",
	})
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string
		Password string
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.userService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
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

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string
		Password string
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.userService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
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

	if err = u.SetLoginToken(ctx, user.Id); err != nil {
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

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "logout success")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
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
	err = u.userService.UpdateNonSensitiveInfo(ctx, domain.User{
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
		Code: 2,
		Msg:  "success",
	})
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		NickName string
		Birthday string
		AboutMe  string
	}
	uid := sessions.Default(ctx).Get(userIdKey).(int64)
	user, err := u.userService.Profile(ctx, uid)
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
		Data: Profile{
			Email:    user.Email,
			Phone:    user.Phone,
			NickName: user.NickName,
			Birthday: user.Birthday.Format(time.DateOnly),
			AboutMe:  user.AboutMe,
		},
	})
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		NickName string
		Birthday string
		AboutMe  string
	}

	claims := ctx.MustGet("user").(*ijwt.UserClaims)

	user, err := u.userService.Profile(ctx, claims.Uid)
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
		Data: Profile{
			Email:    user.Email,
			Phone:    user.Phone,
			NickName: user.NickName,
			Birthday: user.Birthday.Format(time.DateOnly),
			AboutMe:  user.AboutMe,
		},
	})
}

func (u *UserHandler) SendLoginCaptcha(ctx *gin.Context) {
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

	err := u.captchaService.Send(ctx, bizLogin, req.Phone)
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

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone   string `json:"phone"`
		Captcha string `json:"captcha"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := u.captchaService.Verify(ctx, bizLogin, req.Phone, req.Captcha)
	switch err {
	case nil:
		break
	case repository.ErrCaptchaVerifyTooManyTimes:
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "verify too many times, please resend the captcha",
		})
		return
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "captcha invalidated",
		})
		return
	}

	user, err := u.userService.FindOrCreate(ctx, req.Phone)
	if err != nil && err != repository.ErrUserDuplicate {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "captcha validate success",
	})
}
