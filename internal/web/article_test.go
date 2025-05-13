package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"red-feed/internal/service"
	ijwt "red-feed/internal/web/jwt"
	"red-feed/pkg/logger"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestArticleHandler_Publish(t *testing.T) {
	testcases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.ArticleService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				//svc := svcmocks.NewMockUserService(ctrl)
				//svc.EXPECT().SignUp(gomock.Any(), domain.User{
				//	Email:    "123@qq.com",
				//	Password: "hello#world123",
				//}).Return(nil)
				//return svc
				return nil
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
					"confirmPassword": "hello#world123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 拿到user service
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc := tc.mock(ctrl)
			// 暂时用不上 codeSvc
			server := gin.Default()
			// 模拟登录态
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", ijwt.UserClaims{
					Uid: 789,
				})
			})
			artHdl := NewArticleHandler(artSvc, &logger.NopLogger{})
			artHdl.RegisterRoutes(server)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)

			// 因为数据是Json格式，因此要设置请求头中content-type
			req.Header.Set("Content-Type", "application/json")
			// 此时这个req就是非常正确的了

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())

		})
	}
}
