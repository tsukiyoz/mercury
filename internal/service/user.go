/**
 * @author tsukiyo
 * @date 2023-08-11 1:00
 */

package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/internal/domain"
	"webook/internal/repository"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
var ErrInvalidUserOrPassword = errors.New("incorrect account or password")

type UserService struct {
	repo *repository.UserRepository
}

func (s *UserService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return s.repo.Create(ctx, u)
}

func (s *UserService) Login(ctx context.Context, email string, password string) error {
	user, err := s.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNoFound {
		return ErrInvalidUserOrPassword
	}
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return ErrInvalidUserOrPassword
	}
	return nil
}

func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{
		repo: r,
	}
}
