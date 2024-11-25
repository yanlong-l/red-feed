package failover

import (
	"context"
	"errors"
	"fmt"
	"red-feed/internal/service/sms"
)

type FailoverSMSService struct {
	svcs []sms.Service
	idx  uint64
}

func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f *FailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tpl, args, numbers...)
		if err == nil {
			return nil
		}
		// 在这个地方要监控住
		fmt.Println(err)
	}
	return errors.New("全部的服务商都失败了")
}
