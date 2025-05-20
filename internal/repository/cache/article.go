package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"red-feed/internal/domain"
)

type ArticleCache interface {
	SetFirstPage(ctx context.Context, authorId int64, arts []domain.Article) error
	GetFirstPage(ctx context.Context, authorId int64) ([]domain.Article, error)
	DelFirstPage(ctx context.Context, authorId int64) error

	Set(ctx context.Context, art domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)

	SetPub(ctx context.Context, article domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}
type RedisArticleCache struct {
	client redis.Cmdable
}

func NewRedisArticleCache(cmd redis.Cmdable) *RedisArticleCache {
	return &RedisArticleCache{client: cmd}
}

func (c *RedisArticleCache) SetFirstPage(ctx context.Context, authorId int64, arts []domain.Article) error {
	//TODO implement me
	panic("implement me")
}

func (c *RedisArticleCache) GetFirstPage(ctx context.Context, authorId int64) ([]domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

func (c *RedisArticleCache) DelFirstPage(ctx context.Context, authorId int64) error {
	//TODO implement me
	panic("implement me")
}
