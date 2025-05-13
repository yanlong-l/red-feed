package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"red-feed/internal/integration/startup"
	"red-feed/internal/repository/dao"
	ijwt "red-feed/internal/web/jwt"
	"testing"
)

type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	// 初始化db
	s.db = startup.InitTestDB()
	// 初始化测试环境
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})
	artHdl := startup.InitArticleHandler()
	artHdl.RegisterRoutes(s.server)
}

func (s *ArticleTestSuite) TearDownSuite() {
	// 清理测试环境
	// 清空所有数据，并且自增主键恢复到 1
	s.db.Exec("TRUNCATE TABLE articles")
	s.db.Exec("TRUNCATE TABLE publish_articles")
}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testcases := []struct {
		name string
		// 集成测试准备数据
		before func()
		// 集成测试验证数据
		after func()
		// 集成测试的输入
		art Article
		// 预期的HTTP Code
		wantCode int
		// 预期的返回
		wantRes Result[int64]
	}{
		{
			name:   "新建帖子-保存成功",
			before: func() {},
			after:  func() {},
			art: Article{
				Title:   "title1",
				Content: "content1",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
				Msg:  "OK",
			},
		},
		{
			name: "修改已有帖子，并保存",
			before: func() {
				// 提前准备数据
				err := s.db.Create(dao.Article{
					Id:       2,
					Title:    "title1",
					Content:  "content1",
					AuthorId: 123,
					Ctime:    1,
					Utime:    1,
				}).Error
				assert.NoError(t, err)
			},
			after: func() {
				// 数据库验证数据
				var art dao.Article
				ctx := context.Background()
				err := s.db.WithContext(ctx).Find(&art, "id = ?", 2).Error
				assert.NoError(t, err)
				// 验证
				assert.Equal(t, art.Content, "content2")
				assert.Equal(t, art.Title, "title2")
				assert.True(t, art.Utime > 1)
				art.Utime = 0
				assert.Equal(t, art, dao.Article{
					Id:       2,
					Title:    "title2",
					Content:  "content2",
					AuthorId: 123,
					Ctime:    1,
					Utime:    0,
				})
			},
			art: Article{
				Id:      2,
				Title:   "title2",
				Content: "content2",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 2,
				Msg:  "OK",
			},
		},
		{
			name: "修改别人的帖子",
			before: func() {
				// 提前准备数据
				err := s.db.Create(dao.Article{
					Id:       3,
					Title:    "title1",
					Content:  "content1",
					AuthorId: 456,
					Ctime:    1,
					Utime:    1,
				}).Error
				assert.NoError(t, err)
			},
			after: func() {
				// 数据库验证数据
				var art dao.Article
				ctx := context.Background()
				err := s.db.WithContext(ctx).Find(&art, "id = ?", 3).Error
				assert.NoError(t, err)
				assert.Equal(t, art, dao.Article{
					Id:       3,
					Title:    "title1",
					Content:  "content1",
					AuthorId: 456,
					Ctime:    1,
					Utime:    1,
				})
			},
			art: Article{
				Id:      3,
				Title:   "title2",
				Content: "content2",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 先准备数据
			tc.before()
			// 把request body搞出来
			reqBody, err := json.Marshal(tc.art)
			req := httptest.NewRequest("POST", "/articles/edit", bytes.NewBuffer(reqBody))
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			if resp.Code != http.StatusOK {
				return
			}
			assert.Equal(t, tc.wantCode, resp.Code)
			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			t.Log(webRes)
			assert.Equal(t, tc.wantRes, webRes)
			// 清理数据
			tc.after()
		})
	}
	s.TearDownSuite()
}

func (s *ArticleTestSuite) TestPublish() {
	t := s.T()
	testcases := []struct {
		name string
		// 集成测试准备数据
		before func()
		// 集成测试验证数据
		after func()
		// 集成测试的输入
		art Article
		// 预期的HTTP Code
		wantCode int
		// 预期的返回
		wantRes Result[int64]
	}{
		{
			name:   "新建帖子并发表成功",
			before: func() {},
			after:  func() {},
			art: Article{
				Title:   "title1",
				Content: "content1",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
				Msg:  "OK",
			},
		},
		{
			name: "更新已有的帖子并第一次发表",
			before: func() {
				err := s.db.Create(dao.Article{
					Id:       2,
					Title:    "title1",
					Content:  "content1",
					AuthorId: 123,
					Ctime:    1,
					Utime:    1,
				}).Error
				assert.NoError(t, err)
			},
			after: func() {
				var art dao.Article
				err := s.db.WithContext(context.Background()).Find(&art, "id = ?", 2).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 1)
				art.Utime = 1
				assert.Equal(t, art, dao.Article{
					Id:       2,
					Title:    "title2",
					Content:  "content2",
					AuthorId: 123,
					Ctime:    1,
					Utime:    1,
				})
				var pubArt dao.PublishArticle
				err = s.db.WithContext(context.Background()).Find(&pubArt, "id =?", 2).Error
				assert.NoError(t, err)
				assert.True(t, pubArt.Utime > 1)
				pubArt.Utime = 1
				pubArt.Ctime = 1
				assert.Equal(t, pubArt, dao.PublishArticle{
					Article: dao.Article{
						Id:       2,
						Title:    "title2",
						Content:  "content2",
						AuthorId: 123,
						Ctime:    1,
						Utime:    1,
					},
				})
			},
			art: Article{
				Id:      2,
				Title:   "title2",
				Content: "content2",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 2,
				Msg:  "OK",
			},
		},
		{
			name: "更新已有的帖子并重新发表",
			before: func() {
				err := s.db.Create(dao.Article{
					Id:       3,
					Title:    "title3",
					Content:  "content3",
					AuthorId: 123,
					Ctime:    1,
					Utime:    1,
				}).Error
				assert.NoError(t, err)
				err = s.db.Create(dao.PublishArticle{
					dao.Article{
						Id:       3,
						Title:    "title3",
						Content:  "content3",
						AuthorId: 123,
						Ctime:    1,
						Utime:    1,
					},
				}).Error
				assert.NoError(t, err)
			},
			after: func() {
				var art dao.Article
				err := s.db.WithContext(context.Background()).Find(&art, "id = ?", 3).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 1)
				art.Utime = 1
				assert.Equal(t, art, dao.Article{
					Id:       3,
					Title:    "title3-modified",
					Content:  "content3-modified",
					AuthorId: 123,
					Ctime:    1,
					Utime:    1,
				})
				var pubArt dao.PublishArticle
				err = s.db.WithContext(context.Background()).Find(&pubArt, "id =?", 3).Error
				assert.NoError(t, err)
				assert.True(t, pubArt.Utime > 1)
				pubArt.Utime = 1
				pubArt.Ctime = 1
				assert.Equal(t, pubArt, dao.PublishArticle{
					Article: dao.Article{
						Id:       3,
						Title:    "title3-modified",
						Content:  "content3-modified",
						AuthorId: 123,
						Ctime:    1,
						Utime:    1,
					},
				})
			},
			art: Article{
				Id:      3,
				Title:   "title3-modified",
				Content: "content3-modified",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 3,
				Msg:  "OK",
			},
		},
		{
			name: "更新别人的帖子并发表失败",
			before: func() {
				err := s.db.Create(dao.Article{
					Id:       4,
					Title:    "title3",
					Content:  "content3",
					AuthorId: 789,
					Ctime:    1,
					Utime:    1,
				}).Error
				assert.NoError(t, err)
			},
			after: func() {

			},
			art: Article{
				Id:      4,
				Title:   "title1",
				Content: "content1",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 先准备数据
			tc.before()
			// 把request body搞出来
			reqBody, err := json.Marshal(tc.art)
			req := httptest.NewRequest("POST", "/articles/publish", bytes.NewBuffer(reqBody))
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			if resp.Code != http.StatusOK {
				return
			}
			assert.Equal(t, tc.wantCode, resp.Code)
			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			t.Log(webRes)
			assert.Equal(t, tc.wantRes, webRes)
			// 清理数据
			tc.after()
		})
	}
	fmt.Println("ok")
	s.TearDownSuite()
}

func TestArticle(t *testing.T) {
	suite.Run(t, new(ArticleTestSuite))
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
