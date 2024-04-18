package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/tsukaychan/mercury/internal/web/client"
	"net/http"
	"net/http/httptest"
	"testing"

	service2 "github.com/tsukaychan/mercury/interactive/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsukaychan/mercury/internal/domain"
	"github.com/tsukaychan/mercury/internal/service"
	svcmock "github.com/tsukaychan/mercury/internal/service/mocks"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"github.com/tsukaychan/mercury/pkg/logger"
	"go.uber.org/mock/gomock"
)

type Article struct {
	Title   string
	Content string
}

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.ArticleService, service2.InteractiveService)

		req Article

		wantCode   int
		wantResult Result
	}{
		{
			name: "create and publish success",

			mock: func(ctrl *gomock.Controller) (service.ArticleService, service2.InteractiveService) {
				articleSvc := svcmock.NewMockArticleService(ctrl)
				articleSvc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return articleSvc, nil
			},

			req: Article{
				Title:   "my title",
				Content: "my content",
			},

			wantCode: http.StatusOK,
			wantResult: Result{
				Data: float64(1),
			},
		},
		{
			name: "create and publish failed",

			mock: func(ctrl *gomock.Controller) (service.ArticleService, service2.InteractiveService) {
				articleSvc := svcmock.NewMockArticleService(ctrl)
				articleSvc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("publish failed"))
				return articleSvc, nil
			},

			req: Article{
				Title:   "my title",
				Content: "my content",
			},

			wantCode: http.StatusOK,
			wantResult: Result{
				Code: 5,
				Msg:  "internal error",
			},
		},
		// Defensive Programming
		// TODO Modified existing post, published successfully
		// TODO User not found
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", &ijwt.UserClaims{
					Uid: 123,
				})
			})
			atclSvc, intrSvc := tc.mock(ctrl)
			intrCli := client.NewInteractiveLocalAdapter(intrSvc)
			h := NewArticleHandler(atclSvc, intrCli, logger.NewNopLogger())
			h.RegisterRoutes(server)

			reqBody, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var result Result
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
		})
	}
}
