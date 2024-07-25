package main

import (
	"context"
	followv1 "github.com/lazywoo/mercury/api/proto/gen/follow/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func TestFollowGRPCClient(t *testing.T) {
	c, err := grpc.NewClient("localhost:8091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := followv1.NewFollowServiceClient(c)
	resp, err := client.Follow(context.Background(), &followv1.FollowRequest{
		Follower: 23333,
		Followee: 114514,
	})
	require.NoError(t, err)
	t.Log(resp)
}
