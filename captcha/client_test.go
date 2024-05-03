package main

import (
	"context"
	"testing"

	captchav1 "github.com/tsukaychan/mercury/api/proto/gen/captcha/v1"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var target = "etcd:///service/captcha"

func TestCaptchaGRPCClient(t *testing.T) {
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
	client := captchav1.NewCaptchaServiceClient(c)
	resp, err := client.Send(context.Background(), &captchav1.CaptchaSendRequest{
		Biz:   "tpl_id",
		Phone: "12345678900",
	})
	require.NoError(t, err)
	t.Log(resp)
}
