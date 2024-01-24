package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	redismock "github.com/tsukaychan/webook/internal/repository/mocks/cache/redis"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestCaptchaRedisCache_Set(t *testing.T) {
	type in struct {
		ctx     context.Context
		phone   string
		biz     string
		captcha string
	}
	type want struct {
		err error
	}
	type args struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable
		in   in
		want want
	}
	testCases := []args{
		{
			name: "internal error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismock.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-10))
				cmd.EXPECT().Eval(
					gomock.Any(),
					luaSetCaptcha,
					[]string{"phone_captcha:login:18888888888"},
					"123456",
				).Return(res)
				return cmd
			},
			in: in{
				ctx:     context.Background(),
				biz:     "login",
				phone:   "18888888888",
				captcha: "123456",
			},
			want: want{
				err: ErrInternal,
			},
		},
		{
			name: "send to often",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismock.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(
					gomock.Any(),
					luaSetCaptcha,
					[]string{"phone_captcha:login:18888888888"},
					"123456",
				).Return(res)
				return cmd
			},
			in: in{
				ctx:     context.Background(),
				biz:     "login",
				phone:   "18888888888",
				captcha: "123456",
			},
			want: want{
				err: ErrSetCaptchaTooManyTimes,
			},
		},
		{
			name: "redis error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismock.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("redis error"))
				cmd.EXPECT().Eval(
					gomock.Any(),
					luaSetCaptcha,
					[]string{"phone_captcha:login:18888888888"},
					"123456",
				).Return(res)
				return cmd
			},
			in: in{
				ctx:     context.Background(),
				biz:     "login",
				phone:   "18888888888",
				captcha: "123456",
			},
			want: want{
				err: errors.New("redis error"),
			},
		},
		{
			name: "set success",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismock.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(
					gomock.Any(),
					luaSetCaptcha,
					[]string{"phone_captcha:login:18888888888"},
					"123456",
				).Return(res)
				return cmd
			},
			in: in{
				ctx:     context.Background(),
				biz:     "login",
				phone:   "18888888888",
				captcha: "123456",
			},
			want: want{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := NewCaptchaRedisCache(tc.mock(ctrl))
			err := c.Set(tc.in.ctx, tc.in.biz, tc.in.phone, tc.in.captcha)
			assert.Equal(t, tc.want.err, err)
		})
	}
}
