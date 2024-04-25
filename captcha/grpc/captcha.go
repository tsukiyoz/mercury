package grpc

import (
	"context"

	captchav1 "github.com/tsukaychan/mercury/api/proto/gen/captcha/v1"
	"github.com/tsukaychan/mercury/captcha/service"
	"google.golang.org/grpc"
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

func (c *CaptchaServiceServer) Send(ctx context.Context, req *captchav1.CaptchaSendRequest) (*captchav1.CaptchaSendResponse, error) {
	err := c.svc.Send(ctx, req.Biz, req.Phone)
	return &captchav1.CaptchaSendResponse{}, err
}

func (c *CaptchaServiceServer) Verify(ctx context.Context, req *captchav1.VerifyRequest) (*captchav1.VerifyResponse, error) {
	ans, err := c.svc.Verify(ctx, req.Biz, req.Phone, req.Captcha)
	return &captchav1.VerifyResponse{
		Answer: ans,
	}, err
}
