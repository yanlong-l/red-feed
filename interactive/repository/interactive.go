package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"red-feed/interactive/domain"
	"red-feed/interactive/repository/cache"
	"red-feed/interactive/repository/dao"
	"red-feed/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId, uId int64) error
	DecrLike(ctx context.Context, biz string, bizId, uId int64) error
	IncrCollect(ctx context.Context, biz string, bizId, uId, cId int64) error
	DecrCollect(ctx context.Context, biz string, bizId, uId, cId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uId int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uId int64) (bool, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.Logger
}

func (r *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	vals, err := r.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Interactive, domain.Interactive](vals,
		func(idx int, src dao.Interactive) domain.Interactive {
			return r.toDomain(src)
		}), nil
}

func (r *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	return r.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
}

func (r *CachedInteractiveRepository) Liked(ctx context.Context, biz string, bizId, uId int64) (bool, error) {
	_, err := r.dao.GetLikeInfo(ctx, biz, bizId, uId)
	switch err {
	case nil:
		return true, nil
	case dao.ErrDataNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (r *CachedInteractiveRepository) Collected(ctx context.Context, biz string, bizId, uId int64) (bool, error) {
	_, err := r.dao.GetCollectionInfo(ctx, biz, bizId, uId)
	switch err {
	case nil:
		return true, nil
	case dao.ErrDataNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (r *CachedInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 先读缓存
	intr, err := r.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}
	// 缓存没查到，再查数据库
	intrDAO, err := r.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	intr = r.toDomain(intrDAO)
	go func() {
		setErr := r.cache.Set(ctx, biz, bizId, intr)
		if setErr != nil {
			r.l.Error("回写interactive cache失败", logger.String("biz", biz), logger.Error(setErr))
		}
	}()
	return intr, err
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
	err := r.dao.InsertCollectInfo(ctx, biz, bizId, uId, cId)
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

func (r *CachedInteractiveRepository) DecrCollect(ctx context.Context, biz string, bizId, uId, cId int64) error {
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

func (r *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := r.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return r.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (r *CachedInteractiveRepository) toDomain(intrDAO dao.Interactive) domain.Interactive {
	return domain.Interactive{
		BizId:      intrDAO.BizId,
		CollectCnt: intrDAO.CollectCnt,
		LikeCnt:    intrDAO.LikeCnt,
		ReadCnt:    intrDAO.ReadCnt,
	}
}

func NewInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache, l logger.Logger) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}
