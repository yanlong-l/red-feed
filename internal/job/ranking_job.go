package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"red-feed/internal/service"
	"red-feed/pkg/logger"
	"sync"
	"time"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	key       string
	l         logger.Logger
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func NewRankingJob(svc service.RankingService,
	client *rlock.Client,
	l logger.Logger,
	timeout time.Duration) *RankingJob {
	// 根据你的数据量来，如果要是七天内的帖子数量很多，你就要设置长一点
	return &RankingJob{svc: svc,
		timeout:   timeout,
		client:    client,
		key:       "rlock:cron_job:ranking",
		l:         l,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// Run 按时间调度的，三分钟一次
func (r *RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		// 说明你没拿到锁，你得试着拿锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 我可以设置一个比较短的过期时间
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			// 这边没拿到锁，极大概率是别人持有了锁
			return nil
		}
		r.lock = lock
		// 我怎么保证我这里，一直拿着这个锁？？？
		go func() {
			r.localLock.Lock()
			defer r.localLock.Unlock()
			// 自动续约机制
			err1 := lock.AutoRefresh(r.timeout/2, time.Second)
			// 这里说明退出了续约机制
			// 续约失败了怎么办？
			if err1 != nil {
				// 不怎么办
				// 争取下一次，继续抢锁
				r.l.Error("续约失败", logger.Error(err))
			}
			r.lock = nil
			// lock.Unlock(ctx)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
