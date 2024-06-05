package main

import (
	"context"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"

	"google.golang.org/protobuf/types/known/timestamppb"

	articlev1 "github.com/lazywoo/mercury/api/proto/gen/article/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var target = "etcd:///service/article"

func TestArticleGRPCClient(t *testing.T) {
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
	client := articlev1.NewArticleServiceClient(c)
	{
		resp, err := client.List(context.Background(), &articlev1.ListRequest{
			Author: 1,
			Offset: 0,
			Limit:  3,
		})
		require.NoError(t, err)
		t.Log(resp)
	}
	{
		resp, err := client.ListPub(context.Background(), &articlev1.ListPubRequest{
			StartTime: timestamppb.New(time.Now()),
			Offset:    0,
			Limit:     3,
		})
		require.NoError(t, err)
		t.Log(resp)
	}
}
