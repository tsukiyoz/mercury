package service

import (
	"context"
	"testing"
	"time"

	svcmock "github.com/tsukaychan/webook/internal/service/mocks"

	"github.com/tsukaychan/webook/internal/domain"

	"github.com/stretchr/testify/assert"

	"go.uber.org/mock/gomock"
)

func TestRankingTopN(t *testing.T) {
	topNSize := 3
	const batchSize = 5
	now := time.Now()
	limit := 11
	biz := "article"

	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (ArticleService, InteractiveService)

		wantErr   error
		wantAtcls []domain.Article
	}{
		{
			name: "calculate TopN success",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				atclSvc := svcmock.NewMockArticleService(ctrl)
				intrSvc := svcmock.NewMockInteractiveService(ctrl)

				offset := 0
				total := limit
				for total > 0 {
					n := total
					if total > batchSize {
						n = batchSize
					}
					var atcls []domain.Article
					var ids []int64
					intrMap := make(map[int64]domain.Interactive)

					for i := 0; i < n; i++ {
						id := int64(i + 1 + offset)
						atcls = append(atcls, domain.Article{
							Id:    id,
							Utime: now,
							Ctime: now,
						})
						ids = append(ids, id)
						intrMap[id] = domain.Interactive{
							Biz:     biz,
							BizId:   id,
							LikeCnt: int64(limit) - id,
						}
					}

					atclSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), offset, batchSize).
						Return(atcls, nil)

					intrSvc.EXPECT().GetByIds(gomock.Any(), "article", ids).
						Return(intrMap, nil)

					offset += n
					total -= n
				}

				return atclSvc, intrSvc
			},
			wantAtcls: func() []domain.Article {
				atcls := make([]domain.Article, 0, limit)
				for i := 0; i < topNSize; i++ {
					id := int64(i + 1)
					atcls = append(atcls, domain.Article{
						Id:    id,
						Utime: now,
						Ctime: now,
					})
				}
				// slices.Reverse(atcls)
				return atcls
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			atclSvc, intrSvc := tc.mock(ctrl)
			var svc *BatchRankingService
			svc = &BatchRankingService{
				atclSvc:   atclSvc,
				intrSvc:   intrSvc,
				BatchSize: batchSize,
				TopNSize:  topNSize,
				scoreFunc: func(likeCnt int64, utime time.Time) float64 {
					return svc.score(likeCnt, utime)
				},
			}

			atcls, err := svc.rankTopN(ctx)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantAtcls, atcls)
		})
	}
}
