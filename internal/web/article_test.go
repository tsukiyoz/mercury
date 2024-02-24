package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/service"
	svcmock "github.com/tsukaychan/webook/internal/service/mocks"
	ijwt "github.com/tsukaychan/webook/internal/web/jwt"
	"github.com/tsukaychan/webook/pkg/logger"
	"go.uber.org/mock/gomock"
)

type Article struct {
	Title   string
	Content string
}

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.ArticleService, service.InteractiveService)

		req Article

		wantCode   int
		wantResult Result
	}{
		{
			name: "create and publish success",

			mock: func(ctrl *gomock.Controller) (service.ArticleService, service.InteractiveService) {
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

			mock: func(ctrl *gomock.Controller) (service.ArticleService, service.InteractiveService) {
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
			h := NewArticleHandler(atclSvc, intrSvc, logger.NewNopLogger())
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
