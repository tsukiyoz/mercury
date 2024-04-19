package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	interactivev1 "github.com/tsukaychan/mercury/api/proto/gen/interactive/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCClient(t *testing.T) {
	c, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := interactivev1.NewInteractiveServiceClient(c)
	resp, err := client.Get(context.Background(), &interactivev1.GetRequest{
		Biz:   "test",
		BizId: 2,
		Uid:   345,
	})
	require.NoError(t, err)
	t.Log(resp.Interactive)
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
