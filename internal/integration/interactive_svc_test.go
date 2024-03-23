package integration

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tsukaychan/webook/internal/integration/startup"
	"github.com/tsukaychan/webook/internal/repository/dao"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InteractiveSvcTestSuite struct {
	suite.Suite
	db  *gorm.DB
	rdb redis.Cmdable
}

func TestInteractiveService(t *testing.T) {
	suite.Run(t, &InteractiveSvcTestSuite{})
}

func (s *InteractiveSvcTestSuite) SetupSuite() {
	s.db = startup.InitTestDB()
	s.rdb = startup.InitRedis()
}

func (s *InteractiveSvcTestSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `likes`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `favorites`").Error
	assert.NoError(s.T(), err)
}

func (s *InteractiveSvcTestSuite) TestIncrReadCnt() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64

		wantErr error
	}{
		{
			name: "increase db and redis success",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				err := s.db.Create(dao.Interactive{
					Id:          1,
					Biz:         "test",
					BizId:       2,
					ReadCnt:     3,
					FavoriteCnt: 4,
					LikeCnt:     5,
					Ctime:       6,
					Utime:       7,
				}).Error
				assert.NoError(t, err)

				err = s.rdb.HSet(ctx, "interactive:test:2", "read_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				var intr dao.Interactive
				err := s.db.Where("id = ?", 1).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Utime > 7)
				intr.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:          1,
					Biz:         "test",
					BizId:       2,
					ReadCnt:     4,
					FavoriteCnt: 4,
					LikeCnt:     5,
					Ctime:       6,
				}, intr)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "read_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:   "test",
			bizId: 2,
		},
		{
			name: "increase db success, cache failed",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
					DoUpdates: clause.Assignments(map[string]any{
						"read_cnt":     3,
						"favorite_cnt": 4,
						"like_cnt":     5,
						"ctime":        6,
						"utime":        7,
					}),
				}).Create(dao.Interactive{
					Id:          2,
					Biz:         "test",
					BizId:       3,
					ReadCnt:     3,
					FavoriteCnt: 4,
					LikeCnt:     5,
					Ctime:       6,
					Utime:       7,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				var intr dao.Interactive
				err := s.db.WithContext(ctx).Where("id = ?", 2).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Utime > 7)
				intr.Utime = 0
				assert.Equal(t, intr, dao.Interactive{
					Id:          2,
					Biz:         "test",
					BizId:       3,
					ReadCnt:     4,
					FavoriteCnt: 4,
					LikeCnt:     5,
					Ctime:       6,
				})

				cnt, err := s.rdb.Exists(ctx, "interactive:test:3").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},

			biz:   "test",
			bizId: 3,
		},
		{
			name:   "both db and cache has no data and increase success",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				var intr dao.Interactive
				err := s.db.WithContext(ctx).
					Where("biz = ? AND biz_id = ?", "test", 4).
					First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Utime > 0)
				assert.True(t, intr.Ctime > 0)
				assert.True(t, intr.Id > 0)
				intr.Id, intr.Ctime, intr.Utime = 0, 0, 0
				assert.Equal(t, dao.Interactive{
					Biz:     "test",
					BizId:   4,
					ReadCnt: 1,
				}, intr)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:4").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:   "test",
			bizId: 4,
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.IncrReadCnt(context.Background(), tc.biz, tc.bizId)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}
