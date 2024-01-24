package api

import (
	"bytes"
	"errors"
	"github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/domain"
	redismock "github.com/tsukaychan/webook/internal/repository/mocks/cache/redis"
	"github.com/tsukaychan/webook/internal/service"
	svcmock "github.com/tsukaychan/webook/internal/service/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.UserService

		in struct {
			method string
			url    string
			body   []byte
		}
		expect struct {
			code int
			body string
		}
	}{
		{
			name: "invalid params",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/signup",
				body: []byte(`{
				"emails"="test@163.com",
				"passwords": "for.nothing",
				"confirm_passwords": "for.nothing"
			}`)},
			expect: struct {
				code int
				body string
			}{code: http.StatusBadRequest, body: ""},
		},
		{
			name: "email format invalid",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)

				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/signup",
				body: []byte(`{
				"email": "test",
				"password": "for.nothing",
				"confirm_password": "for.nothing"
			}`)},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":4,\"msg\":\"email format invalid\",\"data\":null}",
			},
		},
		{
			name: "passwords doesn't match",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/signup",
				body: []byte(`{
				"email": "test@163.com",
				"password": "for.nothing1",
				"confirm_password": "for.nothing2"
			}`)},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":4,\"msg\":\"passwords doesn't match\",\"data\":null}",
			},
		},
		{
			name: "password format invalid",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/signup",
				body: []byte(`{
				"email": "test@163.com",
				"password": "for",
				"confirm_password": "for"
			}`)},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":4,\"msg\":\"password format invalid\",\"data\":null}",
			},
		},
		{
			name: "signup success",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(service.ErrUserDuplicate)
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/signup",
				body: []byte(`{
				"email": "test@163.com",
				"password": "for.nothing",
				"confirm_password": "for.nothing"
			}`)},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":4,\"msg\":\"the email has been registered\",\"data\":null}",
			},
		},
		{
			name: "internal error",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("internal error"))
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/signup",
				body: []byte(`{
				"email": "test@163.com",
				"password": "for.nothing",
				"confirm_password": "for.nothing"
			}`)},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":5,\"msg\":\"internal error\",\"data\":null}",
			},
		},
		{
			name: "signup success",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(nil)
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/signup",
				body:   []byte(`{"email":"test@163.com","password":"for.nothing","confirm_password":"for.nothing"}`)},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":2,\"msg\":\"success\",\"data\":null}",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()

			h := NewUserHandler(tc.mock(ctrl), nil, nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(tc.in.method, tc.in.url, bytes.NewBuffer(tc.in.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expect.code, resp.Code)
			assert.Equal(t, tc.expect.body, resp.Body.String())
		})
	}
}

func TestUserHandler_LoginJWT(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.UserService

		in struct {
			method string
			url    string
			body   []byte
		}
		expect struct {
			code int
			body string
		}
	}{
		{
			name: "invalid params",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/login",
				body:   []byte(""),
			},
			expect: struct {
				code int
				body string
			}{code: http.StatusBadRequest, body: ""},
		},
		{
			name: "wrong password",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, service.ErrInvalidUserOrPassword)
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/login",
				body:   []byte(`{"email":"tsukiyo6@163.com","password":"bad_password"}`),
			},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":4,\"msg\":\"incorrect account or password\",\"data\":null}",
			},
		},
		{
			name: "internal error in login",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("internal error"))
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/login",
				body:   []byte(`{"email":"tsukiyo6@163.com","password":"for.nothing"}`),
			},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":5,\"msg\":\"internal error\",\"data\":null}",
			},
		},
		{
			name: "login success",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmock.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, nil)
				return userSvc
			},
			in: struct {
				method string
				url    string
				body   []byte
			}{
				method: http.MethodPost,
				url:    "/users/login",
				body:   []byte(`{"email":"tsukiyo6@163.com","password":"for.nothing"}`),
			},
			expect: struct {
				code int
				body string
			}{
				code: http.StatusOK,
				body: "{\"code\":2,\"msg\":\"login success\",\"data\":null}",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()

			h := NewUserHandler(tc.mock(ctrl), nil, jwt.NewRedisJWTHandler(redismock.NewMockCmdable(ctrl)))
			h.RegisterRoutes(server)

			req, err := http.NewRequest(tc.in.method, tc.in.url, bytes.NewBuffer(tc.in.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expect.code, resp.Code)
			assert.Equal(t, tc.expect.body, resp.Body.String())
		})
	}
}
