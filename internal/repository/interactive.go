package repository

import (
	"context"
	"red-feed/internal/domain"
	"red-feed/internal/repository/cache"
	"red-feed/internal/repository/dao"
	"red-feed/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId, uId int64) error
	DecrLike(ctx context.Context, biz string, bizId, uId int64) error
	IncrCollect(ctx context.Context, biz string, bizId, uId, cId int64) error
	DecrCollect(ctx context.Context, biz string, bizId, uId, cId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.Logger
}

func (r *CachedInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 先读缓存
	intr, err := r.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}
	return domain.Interactive{}, err
	// 缓存没查到，再查数据库
	//intr, err = r.dao.Get(ctx, biz, bizId)
}

func (r *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId, uId int64) error {
	// 操作数据库增加点赞计数和用户点赞关系
	err := r.dao.InsertLikeInfo(ctx, biz, bizId, uId)
	if err != nil {
		return err
	}
	go func() {
		// 在缓存中维护住 biz_bizId 的点赞数
		cErr := r.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
		if cErr != nil {
			r.l.Error("IncrLikeCntIfPresent error", logger.Error(cErr))
		}
	}()
	return nil
}

func (r *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId, uId int64) error {
	// 操作数据库减少点赞计数和用户点赞关系
	err := r.dao.DeleteLikeInfo(ctx, biz, bizId, uId)
	if err != nil {
		return err
	}
	go func() {
		// 在缓存中维护住 biz_bizId 的点赞数
		cErr := r.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
		if cErr != nil {
			r.l.Error("DecrLikeCntIfPresent error", logger.Error(cErr))
		}
	}()
	return nil
}

func (r *CachedInteractiveRepository) IncrCollect(ctx context.Context, biz string, bizId, uId, cId int64) error {
	// 操作数据库增加收藏计数和用户收藏关系
	err := r.dao.DeleteCollectInfo(ctx, biz, bizId, uId, cId)
	if err != nil {
		return err
	}
	go func() {
		// 在缓存中维护住 biz_bizId 的点赞数
		cErr := r.cache.DecrCollectCntIfPresent(ctx, biz, bizId)
		if cErr != nil {
			r.l.Error("DecrCollectCntIfPresent error", logger.Error(cErr))
		}
	}()
	return nil
}

func (r *CachedInteractiveRepository) DecrCollect(ctx context.Context, biz string, bizId, uId, cId int64) error {
	// 操作数据库增加收藏计数和用户收藏关系
	err := r.dao.DeleteCollectInfo(ctx, biz, bizId, uId, cId)
	if err != nil {
		return err
	}
	go func() {
		// 在缓存中维护住 biz_bizId 的点赞数
		cErr := r.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
		if cErr != nil {
			r.l.Error("IncrCollectCntIfPresent error", logger.Error(cErr))
		}
	}()
	return nil
}

func (r *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := r.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return r.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func NewInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache, l logger.Logger) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}
