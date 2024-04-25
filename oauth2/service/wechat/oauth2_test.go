//go:build manual

package wechat

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_service_VerifyCode(t *testing.T) {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("no environment variables found WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("no environment variables found WECHAT_APP_SECRET")
	}
	svc := NewService(appId, appSecret)
	res, err := svc.VerifyCode(context.Background(), "111", "state")
	require.NoError(t, err)
	t.Log(res)
}
