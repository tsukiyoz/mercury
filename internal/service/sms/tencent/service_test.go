package tencent

import (
	"context"
	isms "github.com/tsukaychan/webook/internal/service/sms"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

func TestSender(t *testing.T) {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		t.Fatal()
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")

	c, err := sms.NewClient(common.NewCredential(secretId, secretKey),
		"ap-nanjing",
		profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(c, "1400842696", "tsukiyo", nil)

	testCases := []struct {
		name    string
		tplId   string
		params  []isms.ArgVal
		phones  []string
		wantErr error
	}{
		{
			name:  "发送验证码",
			tplId: "1877556",
			params: []isms.ArgVal{
				{
					Val: "123456",
				},
			},
			// 改成你的手机号码
			phones: []string{"13017794139"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplId, tc.params, tc.phones...)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
