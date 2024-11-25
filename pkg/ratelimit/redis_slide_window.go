package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaScript string

type RedisSlidingWindowLimiter struct {
	cmd      redis.Cmdable
	rate     int           // 窗口内的阈值
	interval time.Duration // 窗口大小
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, rate int, interval time.Duration) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		rate:     rate,
		interval: interval,
	}
}

func (r *RedisSlidingWindowLimiter) Limited(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaScript, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}
