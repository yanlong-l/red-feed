package repository

import (
	"context"
	"database/sql"
	"fmt"
	"red-feed/internal/domain"
	"red-feed/internal/repository/cache"
	"red-feed/internal/repository/dao"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type CachedUserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx *gin.Context, phone string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByWechat(ctx context.Context, openID string) (domain.User, error)
}

type UserRepository struct {
	dao   dao.GORMUserDAO
	cache cache.RedisUserCache
}

func NewUserRepository(dao dao.GORMUserDAO, uc cache.RedisUserCache) CachedUserRepository {
	return &UserRepository{
		dao:   dao,
		cache: uc,
	}
}

func (r *UserRepository) FindByWechat(ctx context.Context, openID string) (domain.User, error) {
	u, err := r.dao.FindByWechat(ctx, openID)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserRepository) FindByPhone(ctx *gin.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从缓存查
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		// 必然是有数据，直接返回
		return u, nil
	}
	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = r.entityToDomain(ue)
	// go func() {
	// 	// 加入缓存
	// 	err = r.cache.Set(ctx, u)
	// 	if err != nil {
	// 		fmt.Println("写入缓存失败")
	// 	}
	// }()
	// 加入缓存
	err = r.cache.Set(ctx, u)
	if err != nil {
		fmt.Println("写入缓存失败")
	}
	return u, err
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *UserRepository) domainToEntity(user domain.User) dao.User {
	return dao.User{
		Id:       user.Id,
		Email:    sql.NullString{String: user.Email, Valid: user.Email != ""},
		Phone:    sql.NullString{String: user.Phone, Valid: user.Phone != ""},
		Password: user.Password,
		WechatOpenID: sql.NullString{
			String: user.WechatInfo.OpenID,
			Valid:  user.WechatInfo.OpenID != "",
		},
		WechatUnionID: sql.NullString{
			String: user.WechatInfo.UnionID,
			Valid:  user.WechatInfo.UnionID != "",
		},
		Ctime: user.Ctime.UnixMilli(),
	}
}

func (r *UserRepository) entityToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Password: user.Password,
		Phone:    user.Phone.String,
		WechatInfo: domain.WechatInfo{
			OpenID:  user.WechatOpenID.String,
			UnionID: user.WechatUnionID.String,
		},
		Ctime: time.UnixMilli(user.Ctime),
	}
}
