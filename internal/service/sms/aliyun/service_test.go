package aliyun

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"os"
	"testing"
)

func TestSender(t *testing.T) {
	config := sdk.NewConfig()
	id, secret := os.Getenv("ALIYUN_ACCESS_KEY_ID"), os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	crediential := credentials.NewAccessKeyCredential(id, secret)

	client, err := dysmsapi.NewClientWithOptions("cn-hangzhou", config, crediential)
	if err != nil {
		panic(err)
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.SignName = "阿里云短信测试"
	request.TemplateCode = "SMS_154950909"
	request.PhoneNumbers = "19858810013"
	request.TemplateParam = "{\"code\":\"1234\"}"

	resp, err := client.SendSms(request)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("response is %v\n", resp)
}
