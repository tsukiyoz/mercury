package client

import (
	"context"
	"math/rand"

	"google.golang.org/grpc"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	interactivev1 "github.com/lazywoo/mercury/api/proto/gen/interactive/v1"
)

type InteractiveClient struct {
	remote interactivev1.InteractiveServiceClient
	local  *InteractiveLocalAdapter

	// threshold [1, 100]
	threshold *atomicx.Value[int32]
}

func NewInteractiveClient(remote interactivev1.InteractiveServiceClient,
	local *InteractiveLocalAdapter,
	threshold int32,
) *InteractiveClient {
	threshold = max(1, min(threshold, 100))
	cli := &InteractiveClient{
		remote:    remote,
		local:     local,
		threshold: atomicx.NewValueOf(threshold),
	}
	return cli
}

func (i *InteractiveClient) IncrReadCnt(ctx context.Context, in *interactivev1.IncrReadCntRequest, opts ...grpc.CallOption) (*interactivev1.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in)
}

func (i *InteractiveClient) Like(ctx context.Context, in *interactivev1.LikeRequest, opts ...grpc.CallOption) (*interactivev1.LikeResponse, error) {
	return i.selectClient().Like(ctx, in)
}

func (i *InteractiveClient) CancelLike(ctx context.Context, in *interactivev1.CancelLikeRequest, opts ...grpc.CallOption) (*interactivev1.CancelLikeResponse, error) {
	return i.selectClient().CancelLike(ctx, in)
}

func (i *InteractiveClient) Favorite(ctx context.Context, in *interactivev1.FavoriteRequest, opts ...grpc.CallOption) (*interactivev1.FavoriteResponse, error) {
	return i.selectClient().Favorite(ctx, in)
}

func (i *InteractiveClient) Get(ctx context.Context, in *interactivev1.GetRequest, opts ...grpc.CallOption) (*interactivev1.GetResponse, error) {
	return i.selectClient().Get(ctx, in)
}

func (i *InteractiveClient) GetByIds(ctx context.Context, in *interactivev1.GetByIdsRequest, opts ...grpc.CallOption) (*interactivev1.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in)
}

func (i *InteractiveClient) selectClient() interactivev1.InteractiveServiceClient {
	num := rand.Int31n(100)
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

func (i *InteractiveClient) UpdateThreshold(threshold int32) {
	threshold = max(1, min(threshold, 100))
	i.threshold.Store(threshold)
}
