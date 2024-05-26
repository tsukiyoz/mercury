package test

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ UserServiceServer = (*UserService)(nil)

type UserService struct {
	UnimplementedUserServiceServer
	Name string
}

func NewUserService(name string) *UserService {
	return &UserService{
		Name: name,
	}
}

func (svc *UserService) GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error) {
	// mock db query
	time.Sleep(15 * time.Millisecond)
	return &GetByIDResp{
		User: &User{
			Id:   req.Id,
			Name: "tsukiyo, from " + svc.Name,
		},
	}, nil
}

type FailService struct {
	UnimplementedUserServiceServer
}

func NewFailService() *FailService {
	return &FailService{}
}

func (svc *FailService) GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error) {
	return &GetByIDResp{}, status.Error(codes.Unavailable, "mock service fail")
}
