package main

import (
	"context"
	"testing"

	rankingv1 "github.com/lazywoo/mercury/api/proto/gen/ranking/v1"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var target = "etcd:///service/ranking"

func TestRankingGRPCClient(t *testing.T) {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(t, err)

	rs, err := resolver.NewBuilder(etcdCli)

	opts := []grpc.DialOption{
		grpc.WithResolvers(rs),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	c, err := grpc.NewClient(target, opts...)
	require.NoError(t, err)

	client := rankingv1.NewRankingServiceClient(c)
	//{
	//	resp, err := client.RankTopN(context.Background(), &rankingv1.RankTopNRequest{})
	//	require.NoError(t, err)
	//	t.Log(resp)
	//}
	{
		resp, err := client.TopN(context.Background(), &rankingv1.TopNRequest{})
		require.NoError(t, err)
		t.Log(resp)
	}
}
