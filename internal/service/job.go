package service

import (
	"context"
	"red-feed/internal/domain"
	"red-feed/internal/repository"
	"red-feed/pkg/logger"
	"time"
)

type JobService interface {
	Preempt(ctx context.Context) (domain.Job, error) // Preempt 抢占
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type cronJobService struct {
	repo            repository.JobRepository
	refreshInterval time.Duration
	l               logger.Logger
}

func (js *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	// 试图抢占一个任务
	j, err := js.repo.Preempt(ctx)
	if err != nil {
		return j, err
	}
	// 抢占后，一直刷新 任务的utime, 证明任务还活着
	ticker := time.NewTicker(js.refreshInterval)
	go func() {
		for range ticker.C {
			js.refresh(j.Id)
		}
	}()
	// 定义该任务的cancel func
	j.CancelFunc = func() error {
		//close(ch)
		// 自己在这里释放掉
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return js.repo.Release(ctx, j.Id)
	}
	return j, err
}

func (js *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	next := j.NextTime()
	if next.IsZero() {
		// 没有下一次
		return js.repo.Stop(ctx, j.Id)
	}
	return js.repo.UpdateNextTime(ctx, j.Id, next)
}

func (js *cronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 续约怎么个续法？
	// 更新一下更新时间就可以
	// 比如说我们的续约失败逻辑就是：处于 running 状态，但是更新时间在三分钟以前
	err := js.repo.UpdateUtime(ctx, id)
	if err != nil {
		// 可以考虑立刻重试
		js.l.Error("续约失败",
			logger.Error(err),
			logger.Int64("jid", id))
	}
}
