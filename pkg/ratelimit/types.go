package ratelimit

import "context"

type Limiter interface {
	// key 就是限流器对象
	// bool 表示是否触发限流
	// error 表示限流器本身是否有异常
	Limited(ctx context.Context, key string) (bool, error)
}

