package repository

import (
	"context"
	"red-feed/internal/repository/cache"
	"red-feed/internal/repository/dao"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId, uId int64) error
	DecrLike(ctx context.Context, biz string, id int64) error
	IncrCollectCnt(ctx context.Context, biz string, bizId int64) error
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func (r *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId, uId int64) error {
	// 操作数据库增加阅读计数
	err := r.dao.InsertLikeInfo(ctx, biz, bizId, uId)

}

func (r *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *CachedInteractiveRepository) IncrCollectCnt(ctx context.Context, biz string, bizId int64) error {
	//TODO implement me
	panic("implement me")
}

func NewInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := r.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return r.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}
