package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ijwt "github.com/tsukaychan/webook/internal/api/jwt"
	"github.com/tsukaychan/webook/internal/integration/startup"
	"github.com/tsukaychan/webook/internal/repository/dao"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})

	s.db = startup.InitTestDB()

	articleHdl := startup.InitArticleHandler()
	articleHdl.RegisterRoutes(s.server)
}

func (s *ArticleTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE articles")
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

func (s *ArticleTestSuite) TestSuite() {
	s.T().Log("hello, this is test suite")
}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		article Article

		wantCode   int
		wantResult Result[int64]
	}{
		{
			name:   "create article",
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
					Content:  "this is a content",
					AuthorId: 123,
				}, article)
			},
			article: Article{
				Title:   "my title",
				Content: "this is a content",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Code: 2,
				Msg:  "success",
				Data: 1,
			},
		},
		{
			name: "update article",
			before: func(t *testing.T) {
				err := s.db.Create(dao.Article{
					Id:       2,
					Title:    "my title",
					Content:  "this is a content",
					AuthorId: 123,
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
					Content:  "this is a new content",
					AuthorId: 123,
					Ctime:    123,
				}, article)
			},
			article: Article{
				Id:      2,
				Title:   "my new title",
				Content: "this is a new content",
			},
			wantCode: http.StatusOK,
			wantResult: Result[int64]{
				Code: 2,
				Msg:  "success",
				Data: 2,
			},
		},
		{
			name: "update someone else's article",
			before: func(t *testing.T) {
				err := s.db.Create(dao.Article{
					Id:       3,
					Title:    "my title",
					Content:  "this is a content",
					AuthorId: 789,
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
					Content:  "this is a content",
					AuthorId: 789,
					Ctime:    123,
					Utime:    234,
				}, article)
			},
			article: Article{
				Id:      3,
				Title:   "my new title",
				Content: "this is a new content",
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
			// generate request
			// execute
			// verify result

			tc.before(t)
			reqBody, err := json.Marshal(tc.article)
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

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
