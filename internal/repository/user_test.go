package repository

import (
	"context"
	"database/sql"
	"errors"
	"red-feed/internal/domain"
	"red-feed/internal/repository/cache"
	cachemocks "red-feed/internal/repository/cache/mocks"
	"red-feed/internal/repository/dao"
	daomocks "red-feed/internal/repository/dao/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	now := time.Now()
	// 去除毫秒以外的部分
	now = time.UnixMilli(now.UnixMilli())
	testcases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.GORMUserDAO, cache.RedisUserCache)
		ctx      context.Context
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，查询成功",
			mock: func(ctrl *gomock.Controller) (dao.GORMUserDAO, cache.RedisUserCache) {
				ud := daomocks.NewMockGORMUserDAO(ctrl)
				uc := cachemocks.NewMockRedisUserCache(ctrl)
				uc.EXPECT().Get(context.Background(), int64(123)).Return(domain.User{}, errors.New("no user found"))
				ud.EXPECT().FindById(context.Background(), int64(123)).
					Return(dao.User{
						Id: 123,
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "this is password",
						Phone: sql.NullString{
							String: "15212345678",
							Valid:  true,
						},
						Ctime: now.UnixMilli(),
					}, nil)
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "this is password",
					Phone:    "15212345678",
					Ctime:    now,
				}).Return(nil)
				return ud, uc
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "this is password",
				Phone:    "15212345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存直接命中",
			mock: func(ctrl *gomock.Controller) (dao.GORMUserDAO, cache.RedisUserCache) {
				ud := daomocks.NewMockGORMUserDAO(ctrl)
				uc := cachemocks.NewMockRedisUserCache(ctrl)
				uc.EXPECT().Get(context.Background(), int64(123)).Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "this is password",
					Phone:    "15212345678",
					Ctime:    now,
				}, nil)
				return ud, uc
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "this is password",
				Phone:    "15212345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，且走数据库查询失败",
			mock: func(ctrl *gomock.Controller) (dao.GORMUserDAO, cache.RedisUserCache) {
				ud := daomocks.NewMockGORMUserDAO(ctrl)
				uc := cachemocks.NewMockRedisUserCache(ctrl)
				uc.EXPECT().Get(context.Background(), int64(123)).Return(domain.User{}, errors.New("no user found"))
				ud.EXPECT().FindById(context.Background(), int64(123)).
					Return(dao.User{
					}, errors.New("db error"))
				return ud, uc
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userRepo := NewUserRepository(tc.mock(ctrl))
			u, err := userRepo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)

		})
	}
}
