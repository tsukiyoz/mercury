/**
 * @author tsukiyo
 * @date 2023-08-11 1:00
 */

package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
)

var ErrUserDuplicate = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("incorrect account or password")
var ErrCaptchaSendFrequently = repository.ErrCaptchaSendTooManyTimes

var _ UserService = (*UserServiceV1)(nil)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error
	Profile(ctx context.Context, uid int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
}

type UserServiceV1 struct {
	repo   repository.UserRepository
	logger logger.Logger
}

func NewUserServiceV1(r repository.UserRepository, logger logger.Logger) UserService {
	return &UserServiceV1{
		repo:   r,
		logger: logger,
	}
}

func (svc *UserServiceV1) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserServiceV1) Login(ctx context.Context, email string, password string) (domain.User, error) {
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

func (svc *UserServiceV1) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	user.Email = ""
	user.Phone = ""
	user.Password = ""
	return svc.repo.Update(ctx, user)
}

func (svc *UserServiceV1) Profile(ctx context.Context, uid int64) (domain.User, error) {
	user, err := svc.repo.FindById(ctx, uid)
	return user, err
}

func (svc *UserServiceV1) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNoFound {
		return u, err
	}
	svc.logger.Info("user not registered", logger.String("phone", phone))
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

func (svc *UserServiceV1) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(ctx, info.OpenID)
	if err != repository.ErrUserNoFound {
		return u, err
	}
	u = domain.User{
		WechatInfo: info,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	return svc.repo.FindByWechat(ctx, info.OpenID)
}
