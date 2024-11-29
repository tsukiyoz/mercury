package grpc

import (
	"context"

	"google.golang.org/grpc"

	oauth2v1 "github.com/lazywoo/mercury/api/gen/oauth2/v1"
	"github.com/lazywoo/mercury/internal/oauth2/service/wechat"
)

type OAuth2ServiceServer struct {
	oauth2v1.UnimplementedOauth2ServiceServer
	svc wechat.Service
}

func NewOAuth2ServiceServer(svc wechat.Service) *OAuth2ServiceServer {
	return &OAuth2ServiceServer{
		svc: svc,
	}
}

func (o *OAuth2ServiceServer) Register(server grpc.ServiceRegistrar) {
	oauth2v1.RegisterOauth2ServiceServer(server, o)
}

func (o *OAuth2ServiceServer) AuthURL(ctx context.Context, req *oauth2v1.AuthURLRequest) (*oauth2v1.AuthURLResponse, error) {
	authURL, err := o.svc.AuthURL(ctx, req.GetState())
	if err != nil {
		return nil, err
	}
	return &oauth2v1.AuthURLResponse{
		Url: authURL,
	}, nil
}

func (o *OAuth2ServiceServer) VerifyCode(ctx context.Context, req *oauth2v1.VerifyCodeRequest) (*oauth2v1.VerifyCodeResponse, error) {
	info, err := o.svc.VerifyCode(ctx, req.GetCode())
	if err != nil {
		return nil, err
	}
	return &oauth2v1.VerifyCodeResponse{
		OpenId:  info.OpenID,
		UnionId: info.UnionID,
	}, nil
}
