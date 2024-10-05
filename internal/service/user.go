package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"red-feed/internal/domain"
	"red-feed/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) SignUp(ctx context.Context, u domain.User) error {
	// 由svc来做密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return s.repo.Create(ctx, u)
}

func (s *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
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

func (s *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return s.repo.FindById(ctx, id)
}

func (s *UserService) FindOrCreate(ctx *gin.Context, phone string) (domain.User, error) {
	u, err := s.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		// nil 会进来
		// 不为ErrUserNotFound的也会进来
		return u, err
	}
	// 未找到用户， 创建一个
	err = s.repo.Create(ctx, domain.User{Phone: phone})
	u, err = s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return u, err
}
