package aliyun

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/bytedance/sonic"

	"github.com/tsukiyo/mercury/internal/sms/service"
)

type Service struct {
	client          *dysmsapi.Client
	accessKey       string
	accessKeySecret string
	signName        string
	regionId        string
}

func NewAliyunService(
	accessID,
	accessKeySecret,
	regionId,
	signName string,
) service.Service {
	config := sdk.NewConfig()
	config.WithTimeout(time.Second * 5)
	credential := credentials.NewAccessKeyCredential(accessID, accessKeySecret)
	client, err := dysmsapi.NewClientWithOptions(regionId, config, credential)
	if err != nil {
		panic(err)
	}
	return &Service{
		client:   client,
		signName: signName,
	}
}

func (s *Service) Send(ctx context.Context, tpl string, target string, args []string, values []string) error {
	if len(args) != len(values) {
		return errors.New("invalid input")
	}

	request := dysmsapi.CreateSendSmsRequest()

	request.Scheme = "https"
	request.SignName = s.signName
	request.TemplateCode = tpl
	request.PhoneNumbers = target
	// 参数信息
	params := make(map[string]string, len(args))
	for i, arg := range args {
		params[arg] = values[i]
	}
	bs, err := sonic.Marshal(params)
	if err != nil {
		return err
	}
	request.TemplateParam = string(bs)

	response, err := s.client.SendSms(request)
	if err != nil {
		return err
	}
	fmt.Printf("response is %v\n", response)
	return nil
}
