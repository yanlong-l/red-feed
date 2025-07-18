package learn

import (
	"github.com/robfig/cron/v3"
	"testing"
	"time"
)

func TestCronExpression(t *testing.T) {
	expr := cron.New(cron.WithSeconds())
	expr.AddFunc("@every 3s", func() {
		t.Log("开始长任务")
		//time.Sleep(time.Second * 12)
		t.Log("结束长任务")
	})
	// s 就是秒, m 就是分钟, h 就是小时，d 就是天
	expr.Start()
	// 模拟运行十秒钟
	time.Sleep(time.Second * 10)
	// 发出停止信号，expr 不会调度新的任务，但是也不会中断已经调度了的任务
	stop := expr.Stop()
	t.Log("已经发出停止信号")
	// 这一句会阻塞，等到所有已经调度（正在运行的）结束，才会返回
	<-stop.Done()
	t.Log("彻底结束")
}
