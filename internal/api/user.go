/**
 * @author tsukiyo
 * @date 2023-08-06 12:45
 */

package api

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/internal/domain"
	"webook/internal/service"
)

type UserHandler struct {
	service     *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
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
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "confirm_password doesn't match password")
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "system error"+err.Error())
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "password format invalid")
		return
	}

	err = u.service.SignUp(ctx, domain.User{
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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "req param error",
		})
		return
	}
	err := u.service.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "incorrect account or password")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "internal error")
		return
	}

	ctx.String(http.StatusOK, "login success")
}

func (u *UserHandler) Edit(ctx *gin.Context) {

}

func (u *UserHandler) Profile(ctx *gin.Context) {

}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	server.POST("/user/signup", u.SignUp)
	server.POST("/user/login", u.Login)
	server.POST("/user/edit", u.Edit)
	server.GET("/user/profile", u.Profile)
}

func NewHandler(userService *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = "^(?![a-zA-Z]+$)(?!\\d+$)(?![^\\da-zA-Z\\s]+$).{8,72}$"
	)
	return &UserHandler{
		service:     userService,
		emailExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}
