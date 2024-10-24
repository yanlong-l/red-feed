package service

import (
	"context"
	"errors"
	"red-feed/internal/domain"
	"red-feed/internal/repository"
	repomocks "red-feed/internal/repository/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func Test_userService_Login(t *testing.T) {
	testcases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.CachedUserRepository
		context  context.Context
		email    string
		password string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.CachedUserRepository {
				repo := repomocks.NewMockCachedUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").
					Return(domain.User{Email: "test@example.com", Password: "$2a$10$AACqP5YsOJhgaTHynfsNOO3LpccrInf/lIsnMDID3KTxLElNWuf4m"}, nil)
				return repo
			},
			email:    "test@example.com",
			password: "123",
			wantUser: domain.User{
				Email:    "test@example.com",
				Password: "$2a$10$AACqP5YsOJhgaTHynfsNOO3LpccrInf/lIsnMDID3KTxLElNWuf4m",
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.CachedUserRepository {
				repo := repomocks.NewMockCachedUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "test@example.com",
			password: "123",
			wantUser: domain.User{
			},
			wantErr: ErrInvalidUserOrPassword,
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.CachedUserRepository {
				repo := repomocks.NewMockCachedUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").
					Return(domain.User{Email: "test@example.com", Password: "$2a$10$AACqP5YsOJhgaTHynfsNOO3LpccrInf/lIsnMDID3KTxLElNWuf4m"}, nil)
				return repo
			},
			email:    "test@example.com",
			password: "1234",
			wantUser: domain.User{},
			wantErr: ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.CachedUserRepository {
				repo := repomocks.NewMockCachedUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").
					Return(domain.User{}, errors.New("DB error"))
				return repo
			},
			email:    "test@example.com",
			password: "1234",
			wantUser: domain.User{},
			wantErr: errors.New("DB error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userSvc := NewUserService(tc.mock(ctrl))

			user, err := userSvc.Login(tc.context, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)

		})
	}
}

func Test_Encrypt(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
	t.Log(string(hash))
}
