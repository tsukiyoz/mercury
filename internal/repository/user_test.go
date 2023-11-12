package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"webook/internal/domain"
	user "webook/internal/repository/cache/user"
	"webook/internal/repository/dao"
	cachemock "webook/internal/repository/mocks/cache/user"
	daomock "webook/internal/repository/mocks/dao"
)

func TestUserCachedRepository_FindById(t *testing.T) {
	now := time.UnixMilli(time.Now().UnixMilli())
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (dao.UserDao, user.UserCache)

		in struct {
			ctx context.Context
			id  int64
		}

		want struct {
			user domain.User
			err  error
		}
	}{
		{
			name: "cache hit",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, user.UserCache) {
				userCache, userDao := cachemock.NewMockUserCache(ctrl), daomock.NewMockUserDao(ctrl)
				userCache.EXPECT().Get(gomock.Any(), int64(123)).
					Return(domain.User{
						Id:       123,
						Password: "for.nothing",
						Email:    "test@163.com",
						Phone:    "18888888888",
						CreateAt: now,
						UpdateAt: now,
					}, nil)

				return userDao, userCache
			},

			in: struct {
				ctx context.Context
				id  int64
			}{
				ctx: context.Background(),
				id:  123,
			},

			want: struct {
				user domain.User
				err  error
			}{
				user: domain.User{
					Id:       123,
					Password: "for.nothing",
					Email:    "test@163.com",
					Phone:    "18888888888",
					CreateAt: now,
					UpdateAt: now,
				},
				err: nil,
			},
		},
		{
			name: "cache miss and get data from dao failed",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, user.UserCache) {
				userCache, userDao := cachemock.NewMockUserCache(ctrl), daomock.NewMockUserDao(ctrl)
				userCache.EXPECT().Get(gomock.Any(), int64(123)).
					Return(domain.User{}, user.ErrKeyNotExist)

				userDao.EXPECT().FindById(gomock.Any(), int64(123)).
					Return(dao.User{}, errors.New("db error"))

				return userDao, userCache
			},

			in: struct {
				ctx context.Context
				id  int64
			}{
				ctx: context.Background(),
				id:  123,
			},

			want: struct {
				user domain.User
				err  error
			}{
				user: domain.User{},
				err:  errors.New("db error"),
			},
		},
		{
			name: "cache miss and get data from dao",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, user.UserCache) {
				userCache, userDao := cachemock.NewMockUserCache(ctrl), daomock.NewMockUserDao(ctrl)
				userCache.EXPECT().Get(gomock.Any(), int64(123)).
					Return(domain.User{}, user.ErrKeyNotExist)

				userDao.EXPECT().FindById(gomock.Any(), int64(123)).
					Return(dao.User{
						Id:       123,
						Password: "for.nothing",
						Email: sql.NullString{
							String: "test@163.com",
							Valid:  true,
						},
						Phone: sql.NullString{
							String: "18888888888",
							Valid:  true,
						},
						CreateAt: now.UnixMilli(),
						UpdateAt: now.UnixMilli(),
					}, nil)

				userCache.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Password: "for.nothing",
					Email:    "test@163.com",
					Phone:    "18888888888",
					CreateAt: now,
					UpdateAt: now,
				}).Return(nil)

				return userDao, userCache
			},

			in: struct {
				ctx context.Context
				id  int64
			}{
				ctx: context.Background(),
				id:  123,
			},

			want: struct {
				user domain.User
				err  error
			}{
				user: domain.User{
					Id:       123,
					Password: "for.nothing",
					Email:    "test@163.com",
					Phone:    "18888888888",
					CreateAt: now,
					UpdateAt: now,
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userDao, userCache := tc.mock(ctrl)
			repo := NewUserCachedRepository(userDao, userCache)

			user, err := repo.FindById(tc.in.ctx, tc.in.id)
			assert.Equal(t, tc.want.err, err)
			assert.Equal(t, tc.want.user, user)
			time.Sleep(time.Second)
		})
	}
}
