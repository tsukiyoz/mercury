package main

import (
	"context"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"

	smsv1 "github.com/tsukaychan/mercury/api/proto/gen/sms/v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var target = "etcd:///service/sms"

func TestSmsGRPCClient(t *testing.T) {
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
	// c, err := grpc.NewClient("localhost:8096", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := smsv1.NewSmsServiceClient(c)
	resp, err := client.Send(context.Background(), &smsv1.SmsSendRequest{
		TplId:  "tpl",
		Target: "18888888888",
		Args:   []string{"code"},
		Values: []string{"123456"},
	})
	require.NoError(t, err)
	t.Log(resp)
}
