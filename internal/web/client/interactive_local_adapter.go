package client

import (
	"context"

	"github.com/tsukaychan/mercury/interactive/domain"

	interactivev1 "github.com/tsukaychan/mercury/api/proto/gen/interactive/v1"
	"github.com/tsukaychan/mercury/interactive/service"
	"google.golang.org/grpc"
)

type InteractiveLocalAdapter struct {
	svc service.InteractiveService
}

func NewInteractiveLocalAdapter(svc service.InteractiveService) *InteractiveLocalAdapter {
	return &InteractiveLocalAdapter{svc: svc}
}

func (i *InteractiveLocalAdapter) IncrReadCnt(ctx context.Context, in *interactivev1.IncrReadCntRequest, opts ...grpc.CallOption) (*interactivev1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &interactivev1.IncrReadCntResponse{}, err
}

func (i *InteractiveLocalAdapter) Like(ctx context.Context, in *interactivev1.LikeRequest, opts ...grpc.CallOption) (*interactivev1.LikeResponse, error) {
	err := i.svc.Like(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interactivev1.LikeResponse{}, err
}

func (i *InteractiveLocalAdapter) CancelLike(ctx context.Context, in *interactivev1.CancelLikeRequest, opts ...grpc.CallOption) (*interactivev1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, in.GetBiz(), in.GetUid(), in.GetUid())
	return &interactivev1.CancelLikeResponse{}, err
}

func (i *InteractiveLocalAdapter) Favorite(ctx context.Context, in *interactivev1.FavoriteRequest, opts ...grpc.CallOption) (*interactivev1.FavoriteResponse, error) {
	err := i.svc.Favorite(ctx, in.GetBiz(), in.GetBizId(), in.GetUid(), in.GetFid())
	return &interactivev1.FavoriteResponse{}, err
}

func (i *InteractiveLocalAdapter) Get(ctx context.Context, in *interactivev1.GetRequest, opts ...grpc.CallOption) (*interactivev1.GetResponse, error) {
	res, err := i.svc.Get(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &interactivev1.GetResponse{Interactive: i.toDTO(res)}, nil
}

func (i *InteractiveLocalAdapter) GetByIds(ctx context.Context, in *interactivev1.GetByIdsRequest, opts ...grpc.CallOption) (*interactivev1.GetByIdsResponse, error) {
	if len(in.BizIds) == 0 {
		return &interactivev1.GetByIdsResponse{}, nil
	}
	data, err := i.svc.GetByIds(ctx, in.GetBiz(), in.GetBizIds())
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*interactivev1.Interactive, len(data))
	for k, v := range data {
		res[k] = i.toDTO(v)
	}
	return &interactivev1.GetByIdsResponse{Interactives: res}, nil
}

func (i *InteractiveLocalAdapter) toDTO(intr domain.Interactive) *interactivev1.Interactive {
	return &interactivev1.Interactive{
		Biz:         intr.Biz,
		BizId:       intr.BizId,
		ReadCnt:     intr.ReadCnt,
		LikeCnt:     intr.LikeCnt,
		FavoriteCnt: intr.FavoriteCnt,
		Liked:       intr.Liked,
		Favorited:   intr.Favorited,
	}
}
