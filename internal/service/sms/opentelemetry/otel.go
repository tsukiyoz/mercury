package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/tsukaychan/mercury/internal/service/sms"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewService(svc sms.Service) sms.Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("github.com/tsukaychan/mercury/internal/service/sms/opentelemetry")
	return &Service{
		svc:    svc,
		tracer: tracer,
	}
}

func (s *Service) Send(ctx context.Context, biz string, args []sms.ArgVal, phones ...string) error {
	ctx, span := s.tracer.Start(ctx, "sms_send")
	defer span.End()
	span.SetAttributes(attribute.String("tplId", biz))
	err := s.svc.Send(ctx, biz, args, phones...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
