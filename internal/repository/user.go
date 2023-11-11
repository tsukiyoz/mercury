/**
 * @author tsukiyo
 * @date 2023-08-11 1:24
 */

package repository

import (
	"context"
	"database/sql"
	"log"
	"time"
	"webook/internal/domain"
	cache "webook/internal/repository/cache/user"
	"webook/internal/repository/dao"
)

var ErrUserDuplicate = dao.ErrUserDuplicate
var ErrUserNoFound = dao.ErrUserNotFound

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
}

type UserCachedRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func (r *UserCachedRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *UserCachedRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserCachedRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserCachedRepository) Update(ctx context.Context, user domain.User) error {
	err := r.dao.UpdateNonZeroFields(ctx, r.domainToEntity(user))
	if err != nil {
		return err
	}
	return r.cache.Delete(ctx, user.Id)
}

func (r *UserCachedRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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

func NewUserCachedRepository(dao dao.UserDao, cache cache.UserCache) UserRepository {
	return &UserCachedRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserCachedRepository) domainToEntity(user domain.User) dao.User {
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
		Birthday: sql.NullInt64{
			Int64: user.Birthday.UnixMilli(),
			Valid: !user.Birthday.IsZero(),
		},
		NickName: sql.NullString{
			String: user.NickName,
			Valid:  user.NickName != "",
		},
		AboutMe: sql.NullString{
			String: user.AboutMe,
			Valid:  user.AboutMe != "",
		},
		Password: user.Password,
		CreateAt: user.CreateAt.UnixMilli(),
		UpdateAt: user.UpdateAt.UnixMilli(),
	}
}

func (r *UserCachedRepository) entityToDomain(user dao.User) domain.User {
	var birthday time.Time
	if user.Birthday.Valid {
		birthday = time.UnixMilli(user.Birthday.Int64)
	}
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Phone:    user.Phone.String,
		Birthday: birthday,
		Password: user.Password,
		NickName: user.NickName.String,
		AboutMe:  user.AboutMe.String,
		CreateAt: time.UnixMilli(user.CreateAt),
		UpdateAt: time.UnixMilli(user.UpdateAt),
	}
}
