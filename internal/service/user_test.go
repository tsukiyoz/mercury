package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository"
	repomock "github.com/tsukaychan/webook/internal/repository/mocks"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceV1_Login(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(controller *gomock.Controller) repository.UserRepository
		in   struct {
			ctx      context.Context
			email    string
			password string
		}
		want struct {
			user domain.User
			err  error
		}
	}{
		{
			name: "user not found",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomock.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "test@163.com").Return(domain.User{}, repository.ErrUserNoFound)
				return repo
			},
			in: struct {
				ctx      context.Context
				email    string
				password string
			}{
				ctx:      context.Background(),
				email:    "test@163.com",
				password: "for.nothing",
			},
			want: struct {
				user domain.User
				err  error
			}{
				user: domain.User{},
				err:  ErrInvalidUserOrPassword,
			},
		},
		{
			name: "db error",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomock.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "test@163.com").Return(domain.User{}, errors.New("db error"))
				return repo
			},
			in: struct {
				ctx      context.Context
				email    string
				password string
			}{
				ctx:      context.Background(),
				email:    "test@163.com",
				password: "for.nothing",
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
			name: "incorrect password",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomock.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "test@163.com").Return(domain.User{
					Email:    "test@163.com",
					Password: "$2a$10$qU/QSCQ7MuMOXvMet9Ng2urnLU8X20LYMlsgLY/8FXwfyivlGLGC5",
					Phone:    "18888888888",
					Ctime:    now,
				}, nil)
				return repo
			},
			in: struct {
				ctx      context.Context
				email    string
				password string
			}{
				ctx:      context.Background(),
				email:    "test@163.com",
				password: "for.nothing",
			},
			want: struct {
				user domain.User
				err  error
			}{
				user: domain.User{},
				err:  ErrInvalidUserOrPassword,
			},
		},
		{
			name: "login success",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomock.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "test@163.com").Return(domain.User{
					Email:    "test@163.com",
					Password: "$2a$10$qU/QSCQ7MuMOXvMet9Ng2urnLU8X20LYMlsgLY/8FXwfyivlGLGC6",
					Phone:    "18888888888",
					Ctime:    now,
				}, nil)
				return repo
			},
			in: struct {
				ctx      context.Context
				email    string
				password string
			}{
				ctx:      context.Background(),
				email:    "test@163.com",
				password: "for.nothing",
			},
			want: struct {
				user domain.User
				err  error
			}{
				user: domain.User{
					Email:    "test@163.com",
					Password: "$2a$10$qU/QSCQ7MuMOXvMet9Ng2urnLU8X20LYMlsgLY/8FXwfyivlGLGC6",
					Phone:    "18888888888",
					Ctime:    now,
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl), nil)

			user, err := svc.Login(tc.in.ctx, tc.in.email, tc.in.password)
			assert.Equal(t, tc.want.err, err)
			assert.Equal(t, tc.want.user, user)
		})
	}
}

func TestEncrypted(t *testing.T) {
	password, err := bcrypt.GenerateFromPassword([]byte("for.nothing"), bcrypt.DefaultCost)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(password))
}
