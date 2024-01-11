package tencent

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, number ...string) error {

	request := sms.NewSendSmsRequest()
	request.SetContext(ctx)
	request.SmsSdkAppId = s.appId
	request.SignName = s.signName
	request.TemplateId = ekit.ToPtr(tplId) // 字符串转换为字符串指针
	request.TemplateParamSet = s.toPtrSlice(args)
	request.PhoneNumberSet = s.toPtrSlice(number)
	// 通过client对象调用想要访问的接口，需要传入请求对象
	response, err := s.client.SendSms(request)
	// 处理异常
	if err != nil {
		fmt.Printf("An API error has returned: %s", err)
		return err
	}

	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			// 不可能进来这里
			continue
		}

		status := *statusPtr

		if status.Code == nil || *(status.Code) != "Ok" {
			// 发送失败
			return fmt.Errorf("短信发送失败 code:%s err:%s", *status.Code, *status.Message)
		}

	}
	return nil

}

// 将字符串切片转换为字符串指针切片
func (s *Service) toPtrSlice(args []string) []*string {
	return slice.Map[string, *string](args, func(idx int, src string) *string {
		return &src
	})
}

func NewService(client *sms.Client, appId string, SignName string) *Service {
	return &Service{
		client:   client,
		appId:    &appId,
		signName: &SignName,
	}

}
