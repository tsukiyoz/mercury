package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/lazywoo/mercury/internal/follow/domain"
	"github.com/lazywoo/mercury/internal/follow/service"
	followv1 "github.com/lazywoo/mercury/pkg/api/follow/v1"
	"github.com/lazywoo/mercury/pkg/grpcx"
)

var _ grpcx.Register = (*FollowServiceServer)(nil)

type FollowServiceServer struct {
	followv1.UnimplementedFollowServiceServer
	svc service.FollowService
}

func (f *FollowServiceServer) Register(srv *grpc.Server) {
	followv1.RegisterFollowServiceServer(srv, f)
}

func NewFollowServiceServer(svc service.FollowService) *FollowServiceServer {
	return &FollowServiceServer{
		svc: svc,
	}
}

func (f *FollowServiceServer) Follow(ctx context.Context, request *followv1.FollowRequest) (*followv1.FollowResponse, error) {
	err := f.svc.Follow(ctx, request.Followee, request.Follower)
	return &followv1.FollowResponse{}, err
}

func (f *FollowServiceServer) CancelFollow(ctx context.Context, request *followv1.CancelFollowRequest) (*followv1.CancelFollowResponse, error) {
	err := f.svc.CancelFollow(ctx, request.Followee, request.Follower)
	return &followv1.CancelFollowResponse{}, err
}

func (f *FollowServiceServer) GetFollowee(ctx context.Context, request *followv1.GetFolloweeRequest) (*followv1.GetFolloweeResponse, error) {
	relationList, err := f.svc.GetFollowee(ctx, request.Follower, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}
	res := make([]*followv1.Relation, 0, len(relationList))
	for _, relation := range relationList {
		res = append(res, f.convertRelationToVO(relation))
	}
	return &followv1.GetFolloweeResponse{
		FollowRelation: res,
	}, nil
}

func (f *FollowServiceServer) GetFollower(ctx context.Context, request *followv1.GetFollowerRequest) (*followv1.GetFollowerResponse, error) {
	relationList, err := f.svc.GetFollower(ctx, request.Followee, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}
	res := make([]*followv1.Relation, 0, len(relationList))
	for _, relation := range relationList {
		res = append(res, f.convertRelationToVO(relation))
	}
	return &followv1.GetFollowerResponse{
		FollowRelation: res,
	}, nil
}

func (f *FollowServiceServer) GetRelation(ctx context.Context, request *followv1.GetRelationRequest) (*followv1.GetRelationResponse, error) {
	relation, err := f.svc.GetRelation(ctx, request.Followee, request.Follower)
	if err != nil {
		return nil, err
	}
	return &followv1.GetRelationResponse{
		FollowRelation: f.convertRelationToVO(relation),
	}, nil
}

func (f *FollowServiceServer) GetFollowStatics(ctx context.Context, request *followv1.GetStaticsRequest) (*followv1.GetStaticsResponse, error) {
	statics, err := f.svc.GetStatics(ctx, request.Uid)
	if err != nil {
		return nil, err
	}
	return &followv1.GetStaticsResponse{
		Statics: f.convertStaticsToVO(statics),
	}, nil
}

func (f *FollowServiceServer) convertRelationToVO(relation domain.Relation) *followv1.Relation {
	return &followv1.Relation{
		Follower: relation.Follower,
		Followee: relation.Followee,
	}
}

func (f *FollowServiceServer) convertStaticsToVO(statics domain.Statics) *followv1.Statics {
	return &followv1.Statics{
		FollowerCount: statics.FollowerCount,
		FolloweeCount: statics.FolloweeCount,
	}
}
