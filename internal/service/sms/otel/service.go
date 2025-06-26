package otel

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"red-feed/internal/service/sms"
)

type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewService(svc sms.Service) *Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("gitee.com/geekbang/basic-go/webook/internal/service/sms/opentelemetry")
	return &Service{
		svc:    svc,
		tracer: tracer,
	}
}

func (s *Service) Send(ctx context.Context,
	tpl string,
	args []string,
	numbers ...string) error {
	ctx, span := s.tracer.Start(ctx, "sms_send_"+tpl,
		// 因为我是一个调用短信服务商的客户端
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End(trace.WithStackTrace(true))

	err := s.svc.Send(ctx, tpl, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}

	return err
}
