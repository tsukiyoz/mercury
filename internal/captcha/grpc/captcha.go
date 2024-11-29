package grpc

import (
	"context"

	"google.golang.org/grpc"

	captchav1 "github.com/lazywoo/mercury/api/gen/captcha/v1"
	"github.com/lazywoo/mercury/internal/captcha/service"
)

type CaptchaServiceServer struct {
	captchav1.UnimplementedCaptchaServiceServer
	svc service.CaptchaService
}

func NewCaptchaServiceServer(svc service.CaptchaService) *CaptchaServiceServer {
	return &CaptchaServiceServer{
		svc: svc,
	}
}

func (c *CaptchaServiceServer) Register(server grpc.ServiceRegistrar) {
	captchav1.RegisterCaptchaServiceServer(server, c)
}

func (c *CaptchaServiceServer) Send(ctx context.Context, req *captchav1.SendRequest) (*captchav1.SendResponse, error) {
	err := c.svc.Send(ctx, req.Biz, req.Phone)
	return &captchav1.SendResponse{}, err
}

func (c *CaptchaServiceServer) Verify(ctx context.Context, req *captchav1.VerifyRequest) (*captchav1.VerifyResponse, error) {
	ans, err := c.svc.Verify(ctx, req.Biz, req.Phone, req.Captcha)
	return &captchav1.VerifyResponse{
		Answer: ans,
	}, err
}
