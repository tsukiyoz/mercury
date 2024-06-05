package tencent

import (
	"context"
	"fmt"

	"github.com/lazywoo/mercury/pkg/ratelimit"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId    *string
	signName *string
	client   *sms.Client
	limiter  ratelimit.Limiter
}

const LimitKey = "sms:tencent"

func (s *Service) Send(ctx context.Context, tpl string, args []string, values []string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &tpl
	req.PhoneNumberSet = s.strToStringPtrSlice(args)
	req.TemplateParamSet = s.strToStringPtrSlice(values)

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

func (s *Service) strToStringPtrSlice(values []string) []*string {
	var res []*string
	for i := range values {
		res = append(res, &values[i])
	}
	return res
}

func NewService(client *sms.Client, appId string, signName string, limiter ratelimit.Limiter) *Service {
	return &Service{
		appId:    &appId,
		signName: &signName,
		client:   client,
		limiter:  limiter,
	}
}
