package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tsukaychan/mercury/article/domain"
	"github.com/tsukaychan/mercury/article/repository/dao"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsukaychan/mercury/internal/integration/startup"
	ijwt "github.com/tsukaychan/mercury/internal/web/jwt"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type ArticleGORMTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleGORMTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("user", &ijwt.UserClaims{
			Uid: 123,
		})
	})

	s.db = startup.InitTestDB()

	articleHdl := startup.InitArticleHandler(dao.NewGORMArticleDAO(s.db))
	articleHdl.RegisterRoutes(s.server)
}

func (s *ArticleGORMTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE articles")
	s.db.Exec("TRUNCATE TABLE published_articles")
}

func TestGORMArticle(t *testing.T) {
	suite.Run(t, &ArticleGORMTestSuite{})
}

func (s *ArticleGORMTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		atcl Article

		wantCode   int
		wantResult Result[int64]
	}{
		{
			name:   "create atcl",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				// check db
				var article dao.Article
				err := s.db.Where("id = ?", 1).First(&article).Error

				assert.NoError(t, err)
				assert.True(t, article.Ctime > 0)
				assert.True(t, article.Utime > 0)
				article.Ctime, article.Utime = 0, 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, article)
			},
			atcl: Article{
				Title:   "my title",
				Content: "my content",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "update atcl",
			before: func(t *testing.T) {
				err := s.db.Create(dao.Article{
					Id:       2,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    123,
					Utime:    234,
				}).Error

				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// check db
				var article dao.Article
				err := s.db.Where("id = ?", 2).First(&article).Error

				assert.NoError(t, err)
				assert.True(t, article.Utime > 234)
				article.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "my new title",
					Content:  "my new content",
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					AuthorId: 123,
					Ctime:    123,
				}, article)
			},
			atcl: Article{
				Id:      2,
				Title:   "my new title",
				Content: "my new content",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "update someone else's atcl",
			before: func(t *testing.T) {
				err := s.db.Create(dao.Article{
					Id:       3,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    123,
					Utime:    234,
				}).Error

				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// check db
				var article dao.Article
				err := s.db.Where("id = ?", 3).First(&article).Error

				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    123,
					Utime:    234,
				}, article)
			},
			atcl: Article{
				Id:      3,
				Title:   "my new title",
				Content: "my new content",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "internal error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			reqBody, err := json.Marshal(tc.atcl)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			s.server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}
}

func (s *ArticleGORMTestSuite) TestArticle_Publish() {
	t := s.T()

	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		req Article

		wantCode   int
		wantResult Result[int64]
	}{
		{
			name:   "create and publish",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				var atcl dao.Article
				err := s.db.Where("author_id = ?", 123).First(&atcl).Error
				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				assert.True(t, atcl.Id > 0)
				atcl.Id, atcl.Ctime, atcl.Utime = 0, 0, 0
				assert.Equal(t, dao.Article{
					Title:    "my title",
					Content:  "my content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, atcl)
				var pubAtcl dao.PublishedArticle
				err = s.db.Where("author_id = ?", 123).First(&pubAtcl).Error
				assert.NoError(t, err)
				assert.True(t, pubAtcl.Ctime > 0)
				assert.True(t, pubAtcl.Utime > 0)
				assert.True(t, pubAtcl.Id > 0)
				pubAtcl.Id, pubAtcl.Ctime, pubAtcl.Utime = 0, 0, 0
				assert.Equal(t, dao.PublishedArticle(
					dao.Article{
						Title:    "my title",
						Content:  "my content",
						AuthorId: 123,
						Status:   domain.ArticleStatusPublished.ToUint8(),
					},
				), pubAtcl)
			},
			req: Article{
				Title:   "my title",
				Content: "my content",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "update unpublished and publish",
			before: func(t *testing.T) {
				err := s.db.Create(&dao.Article{
					Id:       2,
					Title:    "my title",
					Content:  "my content",
					Ctime:    234,
					Utime:    456,
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// validate
				var atcl dao.Article
				err := s.db.Where("id = ?", 2).First(&atcl).Error
				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				atcl.Ctime, atcl.Utime = 0, 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "my new title",
					Content:  "my new content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, atcl)

				var pubAtcl dao.PublishedArticle
				err = s.db.Where("id = ?", 2).First(&pubAtcl).Error
				assert.NoError(t, err)
				assert.True(t, pubAtcl.Ctime > 0)
				assert.True(t, pubAtcl.Utime > 0)
				pubAtcl.Ctime, pubAtcl.Utime = 0, 0

				assert.Equal(t, dao.PublishedArticle(
					dao.Article{
						Id:       2,
						Title:    "my new title",
						Content:  "my new content",
						AuthorId: 123,
						Status:   domain.ArticleStatusPublished.ToUint8(),
					},
				), pubAtcl)
			},
			req: Article{
				Id:      2,
				Title:   "my new title",
				Content: "my new content",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "update published and publish",
			before: func(t *testing.T) {
				atcl := dao.Article{
					Id:       3,
					Title:    "my title",
					Content:  "my content",
					Ctime:    234,
					Utime:    456,
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}
				err := s.db.Create(&atcl).Error
				assert.NoError(t, err)
				pubAtcl := dao.PublishedArticle(
					atcl,
				)
				err = s.db.Create(&pubAtcl).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// validate
				var atcl dao.Article
				err := s.db.Where("id = ?", 3).First(&atcl).Error
				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				atcl.Ctime, atcl.Utime = 0, 0
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "my new title",
					Content:  "my new content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, atcl)

				var pubAtcl dao.PublishedArticle
				err = s.db.Where("id = ?", 3).First(&pubAtcl).Error
				assert.NoError(t, err)
				assert.True(t, pubAtcl.Ctime > 0)
				assert.True(t, pubAtcl.Utime > 0)
				pubAtcl.Ctime, pubAtcl.Utime = 0, 0
				assert.Equal(t, dao.PublishedArticle(
					dao.Article{
						Id:       3,
						Title:    "my new title",
						Content:  "my new content",
						AuthorId: 123,
						Status:   domain.ArticleStatusPublished.ToUint8(),
					},
				), pubAtcl)
			},
			req: Article{
				Id:      3,
				Title:   "my new title",
				Content: "my new content",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 3,
			},
		},
		{
			name: "update someone else's article failed",
			before: func(t *testing.T) {
				atcl := dao.Article{
					Id:       4,
					Title:    "my title",
					Content:  "my content",
					Ctime:    234,
					Utime:    456,
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}
				err := s.db.Create(&atcl).Error
				assert.NoError(t, err)

				pubAtcl := dao.PublishedArticle(
					dao.Article{
						Id:       4,
						Title:    "my title",
						Content:  "my content",
						Ctime:    234,
						Utime:    456,
						AuthorId: 789,
						Status:   domain.ArticleStatusPublished.ToUint8(),
					},
				)
				err = s.db.Create(&pubAtcl).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var atcl dao.Article
				err := s.db.Where("id = ?", 4).First(&atcl).Error

				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				atcl.Ctime, atcl.Utime = 0, 0
				assert.Equal(t, dao.Article{
					Id:       4,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, atcl)

				var pubAtcl dao.PublishedArticle
				err = s.db.Where("id = ?", 4).First(&pubAtcl).Error
				assert.NoError(t, err)
				assert.True(t, pubAtcl.Ctime > 0)
				assert.True(t, pubAtcl.Utime > 0)
				pubAtcl.Ctime, pubAtcl.Utime = 0, 0
				assert.Equal(t, dao.PublishedArticle(
					dao.Article{
						Id:       4,
						Title:    "my title",
						Content:  "my content",
						AuthorId: 789,
						Status:   domain.ArticleStatusPublished.ToUint8(),
					},
				), pubAtcl)
			},
			req: Article{
				Id:      4,
				Title:   "my new title",
				Content: "my new content",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "internal error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// no error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}

			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}
}
