package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type Job struct {
	Id       int64
	Name     string // 比如说 ranking
	Cron     string
	Executor string
	// 通用的任务的抽象，我们也不知道任务的具体细节，所以就搞一个 Cfg
	// 具体任务设置具体的值
	Cfg        string
	CancelFunc func() error
}

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom |
	cron.Month | cron.Dow | cron.Descriptor)

func (j Job) NextTime() time.Time {
	// 你怎么算？要根据 cron 表达式来算
	// 可以做成包变量，因为基本不可能变

	s, _ := parser.Parse(j.Cron)
	return s.Next(time.Now())
}
