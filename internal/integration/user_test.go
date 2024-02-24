package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsukaychan/webook/internal/integration/startup"
	"github.com/tsukaychan/webook/internal/web"
	"github.com/tsukaychan/webook/ioc"
)

func TestUserHandler_e2e_SendLoginCaptcha(t *testing.T) {
	server := startup.InitWebServer()
	rdb := ioc.InitRedis()
	type in struct {
		body string
	}
	type want struct {
		code int
		body web.Result
	}
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		in     in
		want   want
	}{
		{
			name: "send success",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				val, err := rdb.GetDel(ctx, "phone_captcha:login:18888888888").Result()
				cancel()
				assert.NoError(t, err)
				assert.True(t, len(val) == 6)
			},
			in: in{
				body: `
{
	"phone": "18888888888"
}
`,
			},
			want: want{
				code: http.StatusOK,
				body: web.Result{
					Code: 2,
					Msg:  "send success",
					Data: nil,
				},
			},
		},
		{
			name: "send too often",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				_, err := rdb.Set(ctx, "phone_captcha:login:18888888888", "123456", time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				val, err := rdb.GetDel(ctx, "phone_captcha:login:18888888888").Result()
				cancel()
				assert.NoError(t, err)
				assert.True(t, "123456" == val)
			},
			in: in{
				body: `
{
	"phone": "18888888888"
}
`,
			},
			want: want{
				code: http.StatusOK,
				body: web.Result{
					Code: 2,
					Msg:  "send too often, please try again later",
					Data: nil,
				},
			},
		},
		{
			name: "internal error",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				_, err := rdb.Set(ctx, "phone_captcha:login:18888888888", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				val, err := rdb.GetDel(ctx, "phone_captcha:login:18888888888").Result()
				cancel()
				assert.NoError(t, err)
				assert.True(t, "123456" == val)
			},
			in: in{
				body: `
{
	"phone": "18888888888"
}
`,
			},
			want: want{
				code: http.StatusOK,
				body: web.Result{
					Code: 5,
					Msg:  "internal error",
					Data: nil,
				},
			},
		},
		{
			name: "invalid phone",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			in: in{
				body: `
{
	"phone": ""
}
`,
			},
			want: want{
				code: http.StatusOK,
				body: web.Result{
					Code: 4,
					Msg:  "please input your phone number",
					Data: nil,
				},
			},
		},
		{
			name: "invalid data",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			in: in{
				body: `
{
	"phone": ""x
}
`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/captcha/send", bytes.NewBuffer([]byte(tc.in.body)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.want.code, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var apiResp web.Result
			err = json.NewDecoder(resp.Body).Decode(&apiResp)
			require.NoError(t, err)

			assert.Equal(t, tc.want.body, apiResp)
			tc.after(t)
		})
	}
}
