package repository

import (
	"context"
	"red-feed/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CodeRepository struct {
	codeCache *cache.CodeCache
}

func NewCodeRepository(cache *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		codeCache: cache,
	}
}

func (cr *CodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return cr.codeCache.Set(ctx, biz, phone, code)
}

func (cr *CodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return cr.codeCache.Verify(ctx, biz, phone, code)
}
