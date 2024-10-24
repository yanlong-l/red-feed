package cache

import (
	"context"
	"errors"
	"red-feed/internal/repository/cache/redismocks"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testcases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		wantErr error
		biz     string
		phone   string
		code    string
	}{
		{
			name: "验证码设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:10086"}, "123456").Return(res)
				return cmd
			},
			wantErr: nil,
			phone:   "10086",
			biz:     "login",
			code:    "123456",
		},
		{
			name: "验证码发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:10086"}, "123456").Return(res)
				return cmd
			},
			wantErr: ErrCodeSendTooMany,
			phone:   "10086",
			biz:     "login",
			code:    "123456",
		},
		{
			name: "Redis 系统错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("redis 系统错误"))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:10086"}, "123456").Return(res)
				return cmd
			},
			wantErr: errors.New("redis 系统错误"),
			phone:   "10086",
			biz:     "login",
			code:    "123456",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cc := NewCodeCache(tc.mock(ctrl))
			err := cc.Set(context.Background(), tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
