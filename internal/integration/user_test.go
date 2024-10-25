package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"red-feed/internal/web"
	"red-feed/ioc"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	// 搞一个gin server出来，这里我们直接使用wire过的
	server := InitWebServer()
	// 把redis搞出来，方便我们准备和清理测试数据
	rdb := ioc.InitRedis()

	testcases := []struct {
		name     string
		before   func(t *testing.T) // 测试前的准备
		after    func(t *testing.T) // 测试后的清理
		reqBody  string
		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {
				// 无需做操作
			},
			after: func(t *testing.T) {
				// 清理一下数据
				err := rdb.Del(context.Background(), "phone_code:login:10086").Err()
				require.NoError(t, err)
			},
			reqBody:  `{"phone":"10086"}`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
		{
			name: "发送太频繁，请稍后再试",
			before: func(t *testing.T) {
				// 准备一下数据
				rdb.Set(context.Background(),"phone_code:login:10086", "123456", time.Minute*9 + time.Second*40)
			},
			after: func(t *testing.T) {
				// 清理一下数据
				err := rdb.Del(context.Background(), "phone_code:login:10086").Err()
				require.NoError(t, err)
			},
			reqBody:  `{"phone":"10086"}`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送太频繁，请稍后再试",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				// 准备一下数据
				rdb.Set(context.Background(),"phone_code:login:10086", "123456", time.Minute*9 + time.Second*40)
			},
			after: func(t *testing.T) {
				// 清理一下数据
				err := rdb.Del(context.Background(), "phone_code:login:10086").Err()
				require.NoError(t, err)
			},
			reqBody:  `{"phone":"10086"}`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送太频繁，请稍后再试",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 先准备数据
			tc.before(t)
			// 把request body搞出来
			req := httptest.NewRequest("POST", "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			var webRes web.Result
			err := json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)
			// 清理数据
			tc.after(t)
		})
	}
}
