package tencent

import (
	"context"
	"fmt"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId    *string
	signName *string
	client   *sms.Client
}

func (s *Service) Send(ctx context.Context, tplId string, params []string, phones ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &tplId
	req.PhoneNumberSet = s.toStringPtrSlice(phones)
	req.TemplateParamSet = s.toStringPtrSlice(params)

	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	// TODO multi error
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("send sms failed, code: %s, reason: %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toStringPtrSlice(values []string) []*string {
	var res []*string
	for i := range values {
		res = append(res, &values[i])
	}
	return res
}

func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		appId:    &appId,
		signName: &signName,
		client:   client,
	}
}
