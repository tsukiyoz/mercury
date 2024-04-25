package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	articlev1 "github.com/tsukaychan/mercury/api/proto/gen/article/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestArticleGRPCClient(t *testing.T) {
	c, err := grpc.NewClient("localhost:8092", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := articlev1.NewArticleServiceClient(c)
	resp, err := client.List(context.Background(), &articlev1.ListRequest{
		Author: 1,
		Offset: 0,
		Limit:  3,
	})
	require.NoError(t, err)
	t.Log(resp)
}
