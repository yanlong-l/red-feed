package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"red-feed/internal/domain"
)

var (
	//go:embed lua/interative_incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldCollectCnt = "collect_cnt"
	fieldLikeCnt    = "like_cnt"
)

//go:generate mockgen -source=./interactive.go -package=cachemocks -destination=mocks/interactive.mock.go InteractiveCache
type InteractiveCache interface {

	// IncrReadCntIfPresent 如果在缓存中有对应的数据，就 +1
	IncrReadCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	// Get 查询缓存中数据
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
}

func NewRedisInteractiveCache(cmd redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{
		client: cmd,
	}
}

type RedisInteractiveCache struct {
	client redis.Cmdable
}

func (c *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return c.client.Eval(ctx, luaIncrCnt,
		[]string{c.key(biz, bizId)},
		fieldReadCnt, 1).Err()
}

func (c *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return c.client.Eval(ctx, luaIncrCnt,
		[]string{c.key(biz, bizId)},
		fieldLikeCnt, 1).Err()
}

func (c *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return c.client.Eval(ctx, luaIncrCnt,
		[]string{c.key(biz, bizId)},
		fieldLikeCnt, -1).Err()
}

func (c *RedisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return c.client.Eval(ctx, luaIncrCnt,
		[]string{c.key(biz, bizId)},
		fieldCollectCnt, 1).Err()
}

func (c *RedisInteractiveCache) DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return c.client.Eval(ctx, luaIncrCnt,
		[]string{c.key(biz, bizId)},
		fieldCollectCnt, -1).Err()
}

func (c *RedisInteractiveCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (c *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	//TODO implement me
	panic("implement me")
}

func (c *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
