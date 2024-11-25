package service

import (
	"context"
	"errors"
	"red-feed/internal/domain"
	"red-feed/internal/repository"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx *gin.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx *gin.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.CachedUserRepository
}

func NewUserService(repo repository.CachedUserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) SignUp(ctx context.Context, u domain.User) error {
	// 由svc来做密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return s.repo.Create(ctx, u)
}

func (s *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := s.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 用户存在，比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// DEBUG
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (s *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return s.repo.FindById(ctx, id)
}

func (s *userService) FindOrCreate(ctx *gin.Context, phone string) (domain.User, error) {
	u, err := s.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		// nil 会进来
		// 不为ErrUserNotFound的也会进来
		return u, err
	}
	// 未找到用户， 创建一个
	err = s.repo.Create(ctx, domain.User{Phone: phone})
	if err != nil {
		return domain.User{}, err
	}
	u, err = s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return u, err
}

func (s *userService) FindOrCreateByWechat(ctx *gin.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	u, err := s.repo.FindByWechat(ctx, wechatInfo.OpenID)
	// 如果找到了，或发生错误，直接返回
	if err != repository.ErrUserNotFound {
		return u, err
	}
	// 如果没有找到，则创建
	u = domain.User{
		WechatInfo: wechatInfo,
	}
	err = s.repo.Create(ctx, u)
	if err != nil {
		return domain.User{}, err
	}
	return s.repo.FindByWechat(ctx, wechatInfo.OpenID)
}
