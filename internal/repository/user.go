/**
 * @author tsukiyo
 * @date 2023-08-11 1:24
 */

package repository

import (
	"log"
	"context"
	"github.com/gin-gonic/gin"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
var ErrUserNoFound = dao.ErrUserNotFound

type UserRepository struct {
	dao   *dao.UserDao
	cache *cache.UserCache
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

func (r *UserRepository) FindById(ctx *gin.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		// 有数据
		return u, nil
	}
	//if err == cache.ErrKeyNotExist {
	//	// 无数据
	//}
	// 缓存出错
	// TODO 数据库限流

	uv, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = domain.User{
		Id:        uv.Id,
		Email:     uv.Email,
		NickName:  uv.NickName,
		Birthday:  uv.Birthday,
		Biography: uv.Biography,
	}

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			log.Printf("cache set failed! reason:%v \n", err)
		}
	}()
	return u, err
}

func NewUserRepository(dao *dao.UserDao, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}
