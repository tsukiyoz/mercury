package grpc

import (
	"context"

	"google.golang.org/grpc"

	smsv1 "github.com/lazywoo/mercury/api/gen/sms/v1"
	"github.com/lazywoo/mercury/internal/sms/service"
)

type SmsServiceServer struct {
	smsv1.UnimplementedSmsServiceServer
	service service.Service
}

func NewSmsServiceServer(svc service.Service) *SmsServiceServer {
	return &SmsServiceServer{
		service: svc,
	}
}

func (s *SmsServiceServer) Register(server grpc.ServiceRegistrar) {
	smsv1.RegisterSmsServiceServer(server, s)
}

func (s *SmsServiceServer) Send(ctx context.Context, req *smsv1.SendRequest) (*smsv1.SendResponse, error) {
	return &smsv1.SendResponse{}, s.service.Send(ctx, req.GetTplId(), req.GetTarget(), req.GetArgs(), req.GetValues())
}
