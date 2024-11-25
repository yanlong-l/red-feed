package ratelimit

import (
	"context"
	"fmt"
	"red-feed/internal/service/sms"
	"red-feed/pkg/ratelimit"
)

type RateLimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRateLimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *RateLimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limited(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候，
		// 可以不限：你的下游很强，业务可用性要求很高，尽量容错策略
		// 包一下这个错误
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if limited {
		return fmt.Errorf("短信服务被限流了")
	}
	return s.svc.Send(ctx, tpl, args, numbers...)
}
