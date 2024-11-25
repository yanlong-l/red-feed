package retryable

import (
	"context"
	"errors"
	"red-feed/internal/service/sms"
)

// 这个要小心并发问题
type Service struct {
	svc sms.Service
	// 重试
	retryCnt int
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	err := s.svc.Send(ctx, tpl, args, numbers...)
	cnt := 1
	for err != nil && cnt < s.retryCnt {
		err = s.svc.Send(ctx, tpl, args, numbers...)
		if err == nil {
			return nil
		}
		cnt++
	}
	return errors.New("重试多次都失败了")
}
