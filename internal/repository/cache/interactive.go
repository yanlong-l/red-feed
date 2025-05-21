package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"red-feed/internal/domain"
	"strconv"
	"time"
)

var (
	//go:embed lua/interative_incr_cnt.lua
	luaIncrCnt string
)

var ErrKeyNotExist = redis.Nil

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
	// 直接使用 HMGet，即便缓存中没有对应的 key，也不会返回 error
	data, err := c.client.HGetAll(ctx, c.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		// 缓存不存在
		return domain.Interactive{}, ErrKeyNotExist
	}

	// 理论上来说，这里不可能有 error
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)

	return domain.Interactive{
		BizId:      bizId,
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
		ReadCnt:    readCnt,
	}, err
}

func (c *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	key := c.key(biz, bizId)
	err := c.client.HMSet(ctx, key,
		fieldLikeCnt, intr.LikeCnt,
		fieldCollectCnt, intr.CollectCnt,
		fieldReadCnt, intr.ReadCnt).Err()
	if err != nil {
		return err
	}
	return c.client.Expire(ctx, key, time.Minute*15).Err()
}

func (c *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
