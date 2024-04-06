package aliyun

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/bytedance/sonic"
	"github.com/tsukaychan/mercury/internal/service/sms"
)

type Service struct {
	client          *dysmsapi.Client
	accessKey       string
	accessKeySecret string
	signName        string
	templateCode    string
	regionId        string
}

func NewAliyunService(
	accessID,
	accessKeySecret,
	regionId,
	signName,
	templateCode string,
) sms.Service {
	config := sdk.NewConfig()
	config.WithTimeout(time.Second * 5)
	credential := credentials.NewAccessKeyCredential(accessID, accessKeySecret)
	client, err := dysmsapi.NewClientWithOptions(regionId, config, credential)
	if err != nil {
		panic(err)
	}
	return &Service{
		client:       client,
		signName:     signName,
		templateCode: templateCode,
	}
}

func (s *Service) Send(ctx context.Context, tpl string, args []sms.ArgVal, phoneNumbers ...string) error {
	request := dysmsapi.CreateSendSmsRequest()

	request.Scheme = "https"
	request.SignName = s.signName
	request.TemplateCode = s.templateCode
	request.PhoneNumbers = strings.Join(phoneNumbers, ",")
	// 参数信息
	tmpMap := make(map[string]string, len(args))
	for _, arg := range args {
		tmpMap[arg.Name] = arg.Val
	}
	// map  转json 字符串
	byteCode, err := sonic.Marshal(tmpMap)
	if err != nil {
		return err
	}
	request.TemplateParam = string(byteCode)

	response, err := s.client.SendSms(request)
	if err != nil {
		return err
	}
	fmt.Printf("response is %v\n", response)
	return nil
}
