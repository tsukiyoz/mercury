/**
 * @author tsukiyo
 * @date 2023-08-06 12:45
 */

package api

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/internal/service"
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
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
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
	return
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

	if err = u.setJWTToken(ctx, user.Id); err != nil {
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
	return
}

func (u *UserHandler) setJWTToken(ctx *gin.Context, userId int64) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
		Uid:          userId,
		RefreshCount: 1,
		UserAgent:    ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedString, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", signedString)
	return nil
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

	claims := ctx.MustGet("user").(*UserClaims)
	err = u.userService.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       claims.Uid,
		NickName: req.Nickname,
		AboutMe:  req.AboutMe,
		Birthday: birthday,
		UpdateAt: time.Now(),
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

	claims := ctx.MustGet("user").(*UserClaims)

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

	if err = u.setJWTToken(ctx, user.Id); err != nil {
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

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/logout", u.Logout)
	ug.POST("/login_sms/captcha/send", u.SendLoginCaptcha)
	ug.POST("/login_sms/captcha/validate", u.LoginSMS)
}

func NewUserHandler(userService service.UserService, captchaService service.CaptchaService) *UserHandler {
	return &UserHandler{
		userService:    userService,
		captchaService: captchaService,
		emailExp:       regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordExp:    regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}
