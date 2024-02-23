package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository"
	articlerepomocks "github.com/tsukaychan/webook/internal/repository/mocks"
	"github.com/tsukaychan/webook/pkg/logger"
	"go.uber.org/mock/gomock"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(controller *gomock.Controller) repository.ArticleRepository

		article domain.Article

		wantId  int64
		wantErr error
	}{
		{
			name: "create and publish success",

			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				atclRepo := articlerepomocks.NewMockArticleRepository(ctrl)

				atclRepo.EXPECT().Sync(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(1), nil)

				return atclRepo
			},

			article: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantId: 1,
		},
		{
			name: "create and publish failed",

			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				atclRepo := articlerepomocks.NewMockArticleRepository(ctrl)

				atclRepo.EXPECT().Sync(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(0), errors.New("mock db error"))

				return atclRepo
			},

			article: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantId:  0,
			wantErr: errors.New("mock db error"),
		},
		{
			name: "update and publish success",

			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				atclRepo := articlerepomocks.NewMockArticleRepository(ctrl)

				atclRepo.EXPECT().Sync(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(2), nil)

				return atclRepo
			},

			article: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
				Status: domain.ArticleStatusPublished,
			},

			wantId: 2,
		},
		{
			name: "update and publish failed",

			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				atclRepo := articlerepomocks.NewMockArticleRepository(ctrl)

				atclRepo.EXPECT().Sync(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(0), errors.New("mock db error"))

				return atclRepo
			},

			article: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
				Status: domain.ArticleStatusPublished,
			},

			wantId:  0,
			wantErr: errors.New("mock db error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			atclRepo := tc.mock(ctrl)

			svc := NewArticleService(atclRepo, logger.NewNopLogger())

			id, err := svc.Publish(context.Background(), tc.article)
			assert.Equal(t, tc.wantId, id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
