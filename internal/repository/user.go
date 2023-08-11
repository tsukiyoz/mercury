/**
 * @author tsukiyo
 * @date 2023-08-11 1:24
 */

package repository

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
var ErrUserNoFound = dao.ErrUserNotFound

type UserRepository struct {
	dao *dao.UserDao
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Create(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
		CreateAt: time.Now().UnixMilli(),
		UpdateAt: time.Now().UnixMilli(),
	})
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}
