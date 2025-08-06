package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/tsukiyo/mercury/internal/user/domain"
	repository2 "github.com/tsukiyo/mercury/internal/user/repository"
	"github.com/tsukiyo/mercury/pkg/logger"
)

var (
	ErrUserDuplicate         = repository2.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("incorrect account or password")
)

var _ UserService = (*userService)(nil)

//go:generate mockgen -source=./user.go -package=svcmocks -destination=mocks/user.mock.go UserService
type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error
	Profile(ctx context.Context, uid int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo   repository2.UserRepository
	logger logger.Logger
}

func NewUserService(r repository2.UserRepository, logger logger.Logger) UserService {
	return &userService{
		repo:   r,
		logger: logger,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, email)
	if err == repository2.ErrUserNoFound {
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

func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	user.Email = ""
	user.Phone = ""
	user.Password = ""
	return svc.repo.Update(ctx, user)
}

func (svc *userService) Profile(ctx context.Context, uid int64) (domain.User, error) {
	user, err := svc.repo.FindById(ctx, uid)
	return user, err
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository2.ErrUserNoFound {
		return u, err
	}
	svc.logger.Info("user not registered", logger.String("phone", phone))
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
		Ctime: time.Now(),
	})
	if err != nil {
		return u, err
	}
	// TODO master-slave delay ?
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(ctx, info.OpenID)
	if err != repository2.ErrUserNoFound {
		return u, err
	}
	u = domain.User{
		WechatInfo: info,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil && err != repository2.ErrUserDuplicate {
		return u, err
	}
	return svc.repo.FindByWechat(ctx, info.OpenID)
}
