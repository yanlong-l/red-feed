package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"red-feed/internal/integration/startup"
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
				"title1",
				"content1",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
				Msg:  "OK",
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

}

func TestArticle(t *testing.T) {
	suite.Run(t, new(ArticleTestSuite))
}

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
