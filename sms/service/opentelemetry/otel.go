package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/tsukaychan/mercury/sms/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	svc    service.Service
	tracer trace.Tracer
}

func NewService(svc service.Service) service.Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("github.com/tsukaychan/mercury/sms/service/opentelemetry/otel.go")
	return &Service{
		svc:    svc,
		tracer: tracer,
	}
}

func (s *Service) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	ctx, span := s.tracer.Start(ctx, "sms_send")
	defer span.End()
	span.SetAttributes(attribute.String("tplId", tpl))
	err := s.svc.Send(ctx, tpl, target, args, values)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
