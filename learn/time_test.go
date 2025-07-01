package learn

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestTimer 相当于只执行一次
func TestTimer(t *testing.T) {
	ticker := time.NewTimer(time.Second * 5)
	defer ticker.Stop()
	go func() {
		for now := range ticker.C {
			fmt.Println(now)
		}
	}()
	time.Sleep(time.Second * 10)
}

// TestTicker 每次时间间隔都会执行
func TestTicker(t *testing.T) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for {
		select {
		case <-ticker.C:
			fmt.Println(time.Now())
		case <-ctx.Done():
			t.Log("超时了，或者被取消了")
			// break 不会退出循环
			goto end
		}
	}
end:
	t.Log("退出循环")

}
