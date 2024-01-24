package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/repository/article"
	articlerepomocks "github.com/tsukaychan/webook/internal/repository/mocks/article"
	"github.com/tsukaychan/webook/pkg/logger"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_articleService_PublishV1(t *testing.T) {
	testCases := []struct {
		name string

		mock func(controller *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository)

		article domain.Article

		wantId  int64
		wantErr error
	}{
		{
			name: "create and publish success",

			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository) {
				articleAuthorRepo := articlerepomocks.NewMockArticleAuthorRepository(ctrl)
				articleReaderRepo := articlerepomocks.NewMockArticleReaderRepository(ctrl)

				articleAuthorRepo.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					}}).Return(int64(1), nil)

				articleReaderRepo.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)

				return articleAuthorRepo, articleReaderRepo
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
			name: "modify and publish success",

			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository) {
				articleAuthorRepo := articlerepomocks.NewMockArticleAuthorRepository(ctrl)
				articleReaderRepo := articlerepomocks.NewMockArticleReaderRepository(ctrl)

				articleAuthorRepo.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					}}).Return(nil)

				articleReaderRepo.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(2), nil)

				return articleAuthorRepo, articleReaderRepo
			},

			article: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},

			wantId: 2,
		},
		// TODO create and publish into production library failed
		// TODO create and publish into online library failed and retry success
		// TODO create and publish into online library failed and retry failed
		// TODO modify and publish into production library failed
		// TODO modify and publish into online library failed and retry success
		// TODO modify and publish into online library failed and retry failed
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo, readerRepo := tc.mock(ctrl)
			svc := NewArticleServiceV1(authorRepo, readerRepo, logger.NewNopLogger())

			id, err := svc.PublishV1(context.Background(), tc.article)
			assert.Equal(t, tc.wantId, id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
