package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"red-feed/interactive/domain"
	"red-feed/interactive/repository"
	"red-feed/pkg/logger"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error                                 // 增加阅读计数
	Like(ctx context.Context, biz string, bizId, uId int64) error                                   // 点赞
	CancelLike(ctx context.Context, biz string, bizId, uId int64) error                             // 取消点赞
	Collect(ctx context.Context, biz string, bizId, uId, cId int64) error                           // 收藏
	CancelCollect(ctx context.Context, biz string, bizId, uId, cId int64) error                     // 取消收藏
	Get(ctx context.Context, biz string, bizId int64, uId int64) (domain.Interactive, error)        // 获取收藏点赞信息
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) // 拿一批文章的interactive信息，用于ranking计算score
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.Logger
}

func (s *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	intrs, err := s.repo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}

func (s *interactiveService) Like(ctx context.Context, biz string, bizId, uId int64) error {
	return s.repo.IncrLike(ctx, biz, bizId, uId)
}

func (s *interactiveService) CancelLike(ctx context.Context, biz string, bizId, uId int64) error {
	return s.repo.DecrLike(ctx, biz, bizId, uId)
}

func (s *interactiveService) Collect(ctx context.Context, biz string, bizId, uId, cId int64) error {
	return s.repo.IncrCollect(ctx, biz, bizId, uId, cId)
}

func (s *interactiveService) CancelCollect(ctx context.Context, biz string, bizId, uId, cId int64) error {
	return s.repo.DecrCollect(ctx, biz, bizId, uId, cId)
}

func (s *interactiveService) Get(ctx context.Context, biz string, bizId int64, uId int64) (domain.Interactive, error) {
	intr, err := s.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		intr.Liked, err = s.repo.Liked(ctx, biz, bizId, uId)
		return err
	})
	eg.Go(func() error {
		intr.Collected, err = s.repo.Collected(ctx, biz, bizId, uId)
		return err
	})
	// 说明是登录过的，补充用户是否点赞或者
	// 新的打印日志的形态 zap 本身就有这种用法
	err = eg.Wait()
	if err != nil {
		// 这个查询失败只需要记录日志就可以，不需要中断执行
		s.l.Error("查询用户是否点赞和收藏的信息失败",
			logger.String("biz", biz),
			logger.Int64("bizId", bizId),
			logger.Int64("uid", uId),
			logger.Error(err))
	}
	return intr, err
}

func (s *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return s.repo.IncrReadCnt(ctx, biz, bizId)
}

func NewInteractiveService(repo repository.InteractiveRepository, l logger.Logger) InteractiveService {
	return &interactiveService{
		repo: repo,
		l:    l,
	}
}
