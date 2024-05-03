package main

import (
	"context"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"

	"github.com/stretchr/testify/require"
	interactivev1 "github.com/tsukaychan/mercury/api/proto/gen/interactive/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var target = "etcd:///service/interactive"

func TestGRPCClient(t *testing.T) {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(t, err)

	rs, err := resolver.NewBuilder(etcdCli)
	require.NoError(t, err)

	opts := []grpc.DialOption{
		grpc.WithResolvers(rs),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	c, err := grpc.Dial(target, opts...)
	require.NoError(t, err)
	client := interactivev1.NewInteractiveServiceClient(c)

	//{
	//	resp, err := client.Get(context.Background(), &interactivev1.GetRequest{
	//		Biz:   "test",
	//		BizId: 2,
	//		Uid:   345,
	//	})
	//	require.NoError(t, err)
	//	t.Log(resp.Interactive)
	//}

	{
		resp, err := client.GetByIds(context.Background(), &interactivev1.GetByIdsRequest{
			Biz:    "article",
			BizIds: []int64{3, 2, 1},
		})
		require.NoError(t, err)
		t.Log(resp.Interactives)
	}
}

func TestDualWrite(t *testing.T) {
	c, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := interactivev1.NewInteractiveServiceClient(c)
	resp, err := client.IncrReadCnt(context.Background(), &interactivev1.IncrReadCntRequest{
		Biz:   "test",
		BizId: 2,
	})
	require.NoError(t, err)
	t.Log(resp.String())
}
