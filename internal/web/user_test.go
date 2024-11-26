package web

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"red-feed/internal/domain"
	"red-feed/internal/service"
	svcmocks "red-feed/internal/service/mocks"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testcases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(nil)
				return svc
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
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				return svc
			},
			reqBody: `
				{
					"email": "123",
					"password": "hello#world123",
					"confirmPassword": "hello#world123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "你的邮箱格式不对",
		},
		{
			name: "密码不一致",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				return svc
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
					"confirmPassword": "hello#world1234"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "两次输入的密码不一致",
		},
		{
			name: "密码格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				return svc
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "123",
					"confirmPassword": "123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "密码必须大于8位，包含数字、特殊字符",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(service.ErrUserDuplicateEmail)
				return svc
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
					"confirmPassword": "hello#world123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(errors.New("mock error"))
				return svc
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
					"confirmPassword": "hello#world123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 拿到user service
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userSvc := tc.mock(ctrl)
			// 暂时用不上 codeSvc
			server := gin.Default()
			userHdl := NewUserHandler(userSvc, nil, nil)
			userHdl.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
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

func TestMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userSvc := svcmocks.NewMockUserService(ctrl)
	userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))

	err := userSvc.SignUp(nil, domain.User{})

	assert.Equal(t, err, errors.New("mock error"))
}
