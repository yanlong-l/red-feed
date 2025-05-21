package service

import (
	"context"
	"red-feed/internal/domain"
	"red-feed/internal/repository"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error               // 增加阅读计数
	Like(ctx context.Context, biz string, bizId, uId int64) error                 // 点赞
	CancelLike(ctx context.Context, biz string, bizId, uId int64) error           // 取消点赞
	Collect(ctx context.Context, biz string, bizId, uId, cId int64) error         // 收藏
	CancelCollect(ctx context.Context, biz string, bizId, uId, cId int64) error   // 取消收藏
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) // 获取收藏点赞信息
}

type interactiveService struct {
	repo repository.InteractiveRepository
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

func (s *interactiveService) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	return s.repo.Get(ctx, biz, bizId)
}

func (s *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return s.repo.IncrReadCnt(ctx, biz, bizId)
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
}
