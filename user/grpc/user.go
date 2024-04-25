package grpc

import (
	"context"

	"github.com/tsukaychan/mercury/user/domain"

	userv1 "github.com/tsukaychan/mercury/api/proto/gen/user/v1"
	"github.com/tsukaychan/mercury/user/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserServiceServer struct {
	userv1.UnimplementedUserServiceServer
	service service.UserService
}

func NewUserServiceServer(svc service.UserService) *UserServiceServer {
	return &UserServiceServer{
		service: svc,
	}
}

func (u *UserServiceServer) Register(server grpc.ServiceRegistrar) {
	userv1.RegisterUserServiceServer(server, u)
}

func (u *UserServiceServer) Signup(ctx context.Context, request *userv1.SignUpRequest) (*userv1.SignUpResponse, error) {
	err := u.service.SignUp(ctx, convertToDomain(request.User))
	return &userv1.SignUpResponse{}, err
}

func (u *UserServiceServer) FindOrCreate(ctx context.Context, request *userv1.FindOrCreateRequest) (*userv1.FindOrCreateResponse, error) {
	user, err := u.service.FindOrCreate(ctx, request.Phone)
	return &userv1.FindOrCreateResponse{
		User: toDTO(user),
	}, err
}

func (u *UserServiceServer) Login(ctx context.Context, request *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	user, err := u.service.Login(ctx, request.GetEmail(), request.GetPassword())
	return &userv1.LoginResponse{
		User: toDTO(user),
	}, err
}

func (u *UserServiceServer) Profile(ctx context.Context, request *userv1.ProfileRequest) (*userv1.ProfileResponse, error) {
	user, err := u.service.Profile(ctx, request.GetId())
	return &userv1.ProfileResponse{
		User: toDTO(user),
	}, err
}

func (u *UserServiceServer) UpdateNonSensitiveInfo(ctx context.Context, request *userv1.UpdateNonSensitiveInfoRequest) (*userv1.UpdateNonSensitiveInfoResponse, error) {
	err := u.service.UpdateNonSensitiveInfo(ctx, convertToDomain(request.GetUser()))
	return &userv1.UpdateNonSensitiveInfoResponse{}, err
}

func (u *UserServiceServer) FindOrCreateByWechat(ctx context.Context, request *userv1.FindOrCreateByWechatRequest) (*userv1.FindOrCreateByWechatResponse, error) {
	user, err := u.service.FindOrCreateByWechat(ctx, domain.WechatInfo{
		OpenID:  request.GetInfo().GetOpenId(),
		UnionID: request.GetInfo().GetUnionId(),
	})
	return &userv1.FindOrCreateByWechatResponse{
		User: toDTO(user),
	}, err
}

func convertToDomain(u *userv1.User) domain.User {
	domainUser := domain.User{}
	if u != nil {
		domainUser.Id = u.GetId()
		domainUser.Email = u.GetEmail()
		domainUser.NickName = u.GetNickName()
		domainUser.Password = u.GetPassword()
		domainUser.Phone = u.GetPhone()
		domainUser.AboutMe = u.GetAboutMe()
		domainUser.Ctime = u.GetCtime().AsTime()
		domainUser.WechatInfo = domain.WechatInfo{
			OpenID:  u.GetWechatInfo().GetOpenId(),
			UnionID: u.GetWechatInfo().GetUnionId(),
		}
	}
	return domainUser
}

func toDTO(user domain.User) *userv1.User {
	vUser := &userv1.User{
		Id:       user.Id,
		Email:    user.Email,
		NickName: user.NickName,
		Password: user.Password,
		Phone:    user.Phone,
		AboutMe:  user.AboutMe,
		Ctime:    timestamppb.New(user.Ctime),
		Birthday: timestamppb.New(user.Birthday),
		WechatInfo: &userv1.WechatInfo{
			OpenId:  user.WechatInfo.OpenID,
			UnionId: user.WechatInfo.UnionID,
		},
	}
	return vUser
}
