/**
 * @author tsukiyo
 * @date 2023-08-11 1:24
 */

package repository

import (
	"context"
	"github.com/gin-gonic/gin"
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
	})
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Edit(ctx *gin.Context, uid int64, nickname string, birthday int64, biography string) error {
	return r.dao.UpdateById(ctx, uid, nickname, birthday, biography)
}

func (r *UserRepository) Profile(ctx *gin.Context, uid int64) (domain.User, error) {
	user, err := r.dao.GetById(ctx, uid)
	if err == ErrUserNoFound {
		return domain.User{}, err
	}
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:        user.Id,
		Email:     user.Email,
		NickName:  user.NickName,
		Birthday:  user.Birthday,
		Biography: user.Biography,
	}, err
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}
