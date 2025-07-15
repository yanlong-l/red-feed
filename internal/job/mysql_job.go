package job

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"red-feed/internal/domain"
	"red-feed/internal/service"
	"red-feed/pkg/logger"
)

type Executor interface {
	Name() string // Executor叫什么
	// Exec ctx 是整个任务调度的上下文
	// 当从 ctx.Done 有信号的时候，就需要考虑结束执行
	// 具体实现来控制
	// 真正去执行一个任务
	Exec(ctx context.Context, j domain.Job) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	f, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("not found func: %s", j.Name)
	}
	return f(ctx, j)
}

// Scheduler 调度器
type Scheduler struct {
	execs   map[string]Executor
	svc     service.JobService
	l       logger.Logger
	limiter *semaphore.Weighted
}
