package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	userv1 "github.com/tsukaychan/mercury/api/proto/gen/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestUserGRPCClient(t *testing.T) {
	c, err := grpc.NewClient("localhost:8091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := userv1.NewUserServiceClient(c)
	resp, err := client.Login(context.Background(), &userv1.LoginRequest{
		Email:    "tsukiyo6@163.com",
		Password: "for.nothing",
	})
	require.NoError(t, err)
	t.Log(resp)
}
