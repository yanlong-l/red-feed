package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"red-feed/internal/service/sms"
	"time"
)

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc sms.Service) *PrometheusDecorator {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "internal_test",
		Subsystem: "red_feed",
		Name:      "sms_resp_time",
		Help:      "统计 SMS 服务的性能数据",
	}, []string{"biz"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}

func (p *PrometheusDecorator) Send(ctx context.Context,
	biz string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.vector.WithLabelValues(biz).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, biz, args, numbers...)
}
