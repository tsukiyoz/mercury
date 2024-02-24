package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsukaychan/webook/internal/domain"
	"github.com/tsukaychan/webook/internal/integration/startup"
	articleDao "github.com/tsukaychan/webook/internal/repository/dao/article"
	ijwt "github.com/tsukaychan/webook/internal/web/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/stretchr/testify/suite"
)

type ArticleMongoTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (s *ArticleMongoTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("user", &ijwt.UserClaims{
			Uid: 123,
		})
	})

	s.mdb = startup.InitMongoDB()
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	err = articleDao.InitCollections(s.mdb)
	if err != nil {
		panic(err)
	}
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")

	articleHdl := startup.InitArticleHandler(articleDao.NewMongoDBDAO(s.mdb, node))
	articleHdl.RegisterRoutes(s.server)
}

func (s *ArticleMongoTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := s.mdb.Collection("articles").DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.mdb.Collection("published_articles").DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func TestMongoArticle(t *testing.T) {
	suite.Run(t, &ArticleMongoTestSuite{})
}

func (s *ArticleMongoTestSuite) TestEdit() {
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
			name:   "create article",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// check db
				var atcl articleDao.Article
				err := s.col.FindOne(ctx, bson.M{"author_id": 123}).Decode(&atcl)

				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				assert.True(t, atcl.Id > 0)
				atcl.Id, atcl.Ctime, atcl.Utime = 0, 0, 0
				assert.Equal(t, articleDao.Article{
					Title:    "my title",
					Content:  "my content",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, atcl)
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
			name: "update article",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				_, err := s.col.InsertOne(ctx, &articleDao.Article{
					Id:       2,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    123,
					Utime:    234,
				})

				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				// check db
				var atcl articleDao.Article
				err := s.col.FindOne(ctx, bson.M{"id": 2}).Decode(&atcl)

				assert.NoError(t, err)
				assert.True(t, atcl.Utime > 234)
				atcl.Utime = 0
				assert.Equal(t, articleDao.Article{
					Id:       2,
					Title:    "my new title",
					Content:  "my new content",
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					AuthorId: 123,
					Ctime:    123,
				}, atcl)
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
			name: "update someone else's article",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				_, err := s.col.InsertOne(ctx, &articleDao.Article{
					Id:       3,
					Title:    "my title",
					Content:  "my content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
					AuthorId: 789,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				// check db
				var atcl articleDao.Article
				err := s.col.FindOne(ctx, bson.M{"id": 3}).Decode(&atcl)

				assert.NoError(t, err)
				assert.Equal(t, articleDao.Article{
					Id:       3,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				}, atcl)
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

			assert.Equal(t, tc.wantResult.Code, result.Code)
			assert.Equal(t, tc.wantResult.Msg, result.Msg)
			if tc.wantResult.Data > 0 {
				assert.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}

func (s *ArticleMongoTestSuite) TestArticle_Publish() {
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				var atcl articleDao.Article
				err := s.col.FindOne(ctx, bson.M{"author_id": 123}).Decode(&atcl)
				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				assert.True(t, atcl.Id > 0)
				atcl.Id, atcl.Ctime, atcl.Utime = 0, 0, 0
				assert.Equal(t, articleDao.Article{
					Title:    "my title",
					Content:  "my content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, atcl)
				var pubAtcl articleDao.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.M{"author_id": 123}).Decode(&pubAtcl)
				assert.NoError(t, err)
				assert.True(t, pubAtcl.Ctime > 0)
				assert.True(t, pubAtcl.Utime > 0)
				assert.True(t, pubAtcl.Id > 0)
				pubAtcl.Id, pubAtcl.Ctime, pubAtcl.Utime = 0, 0, 0
				assert.Equal(t, articleDao.PublishedArticle(
					articleDao.Article{
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				_, err := s.col.InsertOne(ctx, &articleDao.Article{
					Id:       2,
					Title:    "my title",
					Content:  "my content",
					Ctime:    234,
					Utime:    456,
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				// validate
				var atcl articleDao.Article
				err := s.col.FindOne(ctx, bson.M{"id": 2}).Decode(&atcl)
				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				atcl.Ctime, atcl.Utime = 0, 0
				assert.Equal(t, articleDao.Article{
					Id:       2,
					Title:    "my new title",
					Content:  "my new content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, atcl)

				var pubAtcl articleDao.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.M{"id": 2, "author_id": 123}).Decode(&pubAtcl)
				assert.NoError(t, err)
				assert.True(t, pubAtcl.Ctime > 0)
				assert.True(t, pubAtcl.Utime > 0)
				pubAtcl.Ctime, pubAtcl.Utime = 0, 0
				assert.Equal(t, articleDao.PublishedArticle(
					articleDao.Article{
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				atcl := articleDao.Article{
					Id:       3,
					Title:    "my title",
					Content:  "my content",
					Ctime:    234,
					Utime:    456,
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}
				_, err := s.col.InsertOne(ctx, &atcl)
				assert.NoError(t, err)
				pubAtcl := articleDao.PublishedArticle(
					atcl,
				)
				_, err = s.liveCol.InsertOne(ctx, &pubAtcl)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				// validate
				var atcl articleDao.Article
				err := s.col.FindOne(ctx, bson.M{"id": 3}).Decode(&atcl)
				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				atcl.Ctime, atcl.Utime = 0, 0
				assert.Equal(t, articleDao.Article{
					Id:       3,
					Title:    "my new title",
					Content:  "my new content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, atcl)

				var pubAtcl articleDao.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.M{"author_id": 123}).Decode(&pubAtcl)
				assert.NoError(t, err)
				assert.True(t, pubAtcl.Ctime > 0)
				assert.True(t, pubAtcl.Utime > 0)
				pubAtcl.Ctime, pubAtcl.Utime = 0, 0
				assert.Equal(t, articleDao.PublishedArticle(
					articleDao.Article{
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				atcl := articleDao.Article{
					Id:       4,
					Title:    "my title",
					Content:  "my content",
					Ctime:    234,
					Utime:    456,
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}
				_, err := s.col.InsertOne(ctx, &atcl)
				assert.NoError(t, err)

				pubAtcl := articleDao.PublishedArticle(
					articleDao.Article{
						Id:       4,
						Title:    "my title",
						Content:  "my content",
						Ctime:    234,
						Utime:    456,
						AuthorId: 789,
						Status:   domain.ArticleStatusPublished.ToUint8(),
					},
				)
				_, err = s.liveCol.InsertOne(ctx, &pubAtcl)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				var atcl articleDao.Article
				err := s.col.FindOne(ctx, bson.M{"id": 4}).Decode(&atcl)

				assert.NoError(t, err)
				assert.True(t, atcl.Ctime > 0)
				assert.True(t, atcl.Utime > 0)
				atcl.Ctime, atcl.Utime = 0, 0
				assert.Equal(t, articleDao.Article{
					Id:       4,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, atcl)

				var pubAtcl articleDao.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.M{"id": 4}).Decode(&pubAtcl)
				assert.NoError(t, err)
				assert.True(t, pubAtcl.Ctime > 0)
				assert.True(t, pubAtcl.Utime > 0)
				pubAtcl.Ctime, pubAtcl.Utime = 0, 0
				assert.Equal(t, articleDao.PublishedArticle(
					articleDao.Article{
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
			assert.Equal(t, tc.wantResult.Code, result.Code)
			assert.Equal(t, tc.wantResult.Msg, result.Msg)
			if tc.wantResult.Data > 0 {
				assert.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}
