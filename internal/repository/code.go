package repository

import (
	"context"
	"red-feed/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CachedCodeRepository interface {
	Store(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error) 
}

type CodeRepository struct {
	codeCache cache.RedisCodeCache
}

func NewCodeRepository(cache cache.RedisCodeCache) CachedCodeRepository {
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
