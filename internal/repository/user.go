/**
 * @author tsukiyo
 * @date 2023-08-11 1:24
 */

package repository

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var ErrUserDuplicate = dao.ErrUserDuplicate
var ErrUserNoFound = dao.ErrUserNotFound

type UserRepository struct {
	dao   *dao.UserDao
	cache *cache.UserCache
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Create(ctx, r.domainToEntity(u))
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
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

	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = r.entityToDomain(ue)

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

func (r *UserRepository) domainToEntity(user domain.User) dao.User {
	return dao.User{
		Id: user.Id,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
		CreateAt: user.CreateAt.UnixMilli(),
	}
}

func (r *UserRepository) entityToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Phone:    user.Phone.String,
		Password: user.Password,
		CreateAt: time.UnixMilli(user.CreateAt),
	}
}
