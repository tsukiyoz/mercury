/**
 * @author tsukiyo
 * @date 2023-08-11 1:00
 */

package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
)

var ErrUserDuplicate = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("incorrect account or password")
var ErrCaptchaSendFrequently = repository.ErrCaptchaSendTooManyTimes

type UserService struct {
	repo *repository.UserRepository
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNoFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return user, nil
}

func (svc *UserService) Edit(ctx *gin.Context, uid int64, nickname string, birthday int64, biography string) error {
	return svc.repo.Edit(ctx, uid, nickname, birthday, biography)
}

func (svc *UserService) Profile(ctx *gin.Context, uid int64) (domain.User, error) {
	user, err := svc.repo.FindById(ctx, uid)
	return user, err
}

func (svc *UserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNoFound {
		return u, err
	}
	err = svc.repo.Create(ctx, domain.User{
		Phone:    phone,
		CreateAt: time.Now(),
	})
	if err != nil {
		return u, err
	}
	// TODO master-slave delay ?
	return svc.repo.FindByPhone(ctx, phone)
}

func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{
		repo: r,
	}
}
