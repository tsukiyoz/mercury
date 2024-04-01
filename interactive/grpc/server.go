package grpc

import (
	"context"

	"github.com/tsukaychan/webook/interactive/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tsukaychan/webook/api/proto/gen/interactive/v1"
	"github.com/tsukaychan/webook/interactive/service"
)

type InteractiveServiceServer struct {
	interactivev1.UnimplementedInteractiveServiceServer
	svc service.InteractiveService
}

func (srv *InteractiveServiceServer) toDTO(intr domain.Interactive) *interactivev1.Interactive {
	return &interactivev1.Interactive{
		Biz:         intr.Biz,
		BizId:       intr.BizId,
		ReadCnt:     intr.BizId,
		LikeCnt:     intr.LikeCnt,
		FavoriteCnt: intr.FavoriteCnt,
		Liked:       intr.Liked,
		Favorited:   intr.Favorited,
	}
}

func (srv *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *interactivev1.IncrReadCntRequest) (*interactivev1.IncrReadCntResponse, error) {
	err := srv.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	if err != nil {
		return nil, err
	}
	return &interactivev1.IncrReadCntResponse{}, nil
}

func (srv *InteractiveServiceServer) Like(ctx context.Context, request *interactivev1.LikeRequest) (*interactivev1.LikeResponse, error) {
	if request.GetUid() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "uid invalid")
	}
	err := srv.svc.Like(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &interactivev1.LikeResponse{}, nil
}

func (srv *InteractiveServiceServer) CancelLike(ctx context.Context, request *interactivev1.CancelLikeRequest) (*interactivev1.CancelLikeResponse, error) {
	err := srv.svc.CancelLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &interactivev1.CancelLikeResponse{}, nil
}

func (srv *InteractiveServiceServer) Favorite(ctx context.Context, request *interactivev1.FavoriteRequest) (*interactivev1.FavoriteResponse, error) {
	err := srv.svc.Favorite(ctx, request.GetBiz(), request.GetBizId(), request.GetUid(), request.GetFid())
	if err != nil {
		return nil, err
	}
	return &interactivev1.FavoriteResponse{}, nil
}

func (srv *InteractiveServiceServer) Get(ctx context.Context, request *interactivev1.GetRequest) (*interactivev1.GetResponse, error) {
	intr, err := srv.svc.Get(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &interactivev1.GetResponse{
		Interactive: srv.toDTO(intr),
	}, nil
}

func (srv *InteractiveServiceServer) GetByIds(ctx context.Context, request *interactivev1.GetByIdsRequest) (*interactivev1.GetByIdsResponse, error) {
	res, err := srv.svc.GetByIds(ctx, request.GetBiz(), request.GetBizIds())
	if err != nil {
		return nil, err
	}

	m := make(map[int64]*interactivev1.Interactive, len(res))
	for _, intr := range res {
		m[intr.BizId] = srv.toDTO(intr)
	}

	return &interactivev1.GetByIdsResponse{Interactives: m}, nil
}
