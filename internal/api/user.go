/**
 * @author tsukiyo
 * @date 2023-08-06 12:45
 */

package api

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
)

type UserHandler struct {
	userService    *service.UserService
	captchaService *service.CaptchaService
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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "signup failed",
		})
		return
	}

	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "system error"+err.Error())
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "email format invalid")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "confirm_password doesn't match password")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "system error: "+err.Error())
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "password format invalid")
		return
	}

	err = u.userService.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "internal error")
		return
	}
	ctx.String(http.StatusOK, "signup success")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string
		Password string
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "req param error")
		return
	}
	user, err := u.userService.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "incorrect account or password")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "internal error")
		return
	}

	ss := sessions.Default(ctx)
	ss.Set("user_id", user.Id)
	ss.Options(sessions.Options{
		Secure:   true,
		HttpOnly: true,
		MaxAge:   30,
	})
	ss.Save()

	ctx.String(http.StatusOK, "login success")
	return
}

func (u *UserHandler) LoginWithJwt(ctx *gin.Context) {
	type LoginReq struct {
		Email    string
		Password string
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "req param error")
		return
	}
	user, err := u.userService.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "incorrect account or password")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "internal error")
		return
	}

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
		Uid:          user.Id,
		RefreshCount: 1,
		UserAgent:    ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedString, err := token.SignedString([]byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "internal error")
	}
	ctx.Header("x-jwt-token", signedString)
	ctx.String(http.StatusOK, "login success")
	return
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
		Nickname  string
		Birthday  string
		Biography string
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "req param error")
		return
	}
	uid, ok := sessions.Default(ctx).Get("user_id").(int64)
	if !ok {
		ctx.String(http.StatusUnauthorized, "no user login")
		return
	}
	ctx.String(http.StatusOK, fmt.Sprintf("%v %v \n", uid, req))
	birthdayTime, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "req param invalid")
		return
	}
	if err := u.userService.Edit(ctx, uid, req.Nickname, birthdayTime.UnixMilli(), req.Biography); err != nil {
		ctx.String(http.StatusOK, "internal error")
		return
	}
	ctx.String(http.StatusOK, "edit success")
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	uid, ok := sessions.Default(ctx).Get("user_id").(int64)
	if !ok {
		ctx.String(http.StatusUnauthorized, "no user login")
		return
	}
	profile, err := u.userService.Profile(ctx, uid)
	if err != nil {
		ctx.String(http.StatusOK, "internal error")
		return
	}
	ctx.String(http.StatusOK, fmt.Sprintf("%v\n", profile))
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, _ := ctx.Get("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "internal error")
	}

	profile, err := u.userService.Profile(ctx, claims.Uid)
	if err != nil {
		ctx.String(http.StatusOK, "internal error")
		return
	}
	ctx.String(http.StatusOK, fmt.Sprintf("%v\n", profile))
}

func (u *UserHandler) SendLoginCaptcha(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	const biz = "login"
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	err := u.captchaService.Send(ctx, biz, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "internal error",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "send success",
	})
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/user")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginWithJwt)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/logout", u.Logout)
	ug.POST("/login_sms/captchaService/send", u.SendLoginCaptcha)
	//ug.POST("/login/captchaService", u.LoginSMS)
}

func NewHandler(userService *service.UserService, captchaService *service.CaptchaService) *UserHandler {
	const (
		emailRegexPattern    = "[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[\\w](?:[\\w-]*[\\w])?"
		passwordRegexPattern = "^(?![a-zA-Z]+$)(?!\\d+$)(?![^\\da-zA-Z\\s]+$).{8,72}$"
	)
	return &UserHandler{
		userService:    userService,
		captchaService: captchaService,
		emailExp:       regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordExp:    regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid          int64
	RefreshCount int64
	UserAgent    string
}
