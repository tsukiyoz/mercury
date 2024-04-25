package grpc

import (
	"context"

	"google.golang.org/grpc"

	smsv1 "github.com/tsukaychan/mercury/api/proto/gen/sms/v1"
	"github.com/tsukaychan/mercury/sms/service"
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

func (s *SmsServiceServer) Send(ctx context.Context, req *smsv1.SmsSendRequest) (*smsv1.SmsSendResponse, error) {
	return &smsv1.SmsSendResponse{}, s.service.Send(ctx, req.GetTplId(), req.GetTarget(), req.GetArgs(), req.GetValues())
}
