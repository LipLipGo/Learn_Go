package tencent

import (
	"Learn_Go/webook/pkg/limiter"
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
	limiter  limiter.Limiter
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	// 下面的这个限流实现：
	// 从功能上讲，没有问题；从扩展性上讲，全是问题；从无侵入式上讲，更加是问题
	// 如果将来我有别的短信服务商，别的短信服务商也需要限流；如果由别的类似的功能，个需要修改这个方法，改来改去就堆成了屎山
	//limited, err := s.limiter.Limit(ctx, "tencent_sms_service")
	//if err != nil {
	//	return err
	//}
	//if limited {
	//	return errors.New("触发了限流")
	//}
	request := sms.NewSendSmsRequest()
	request.SetContext(ctx)
	request.SmsSdkAppId = s.appId
	request.SignName = s.signName
	request.TemplateId = ekit.ToPtr(tplId) // 字符串转换为字符串指针
	request.TemplateParamSet = s.toPtrSlice(args)
	request.PhoneNumberSet = s.toPtrSlice(number)
	// 通过client对象调用想要访问的接口，需要传入请求对象
	response, err := s.client.SendSms(request)

	// 可以将请求和响应DEBUG一下，查看一下数据有没有错误
	zap.L().Debug("调用腾讯短信服务", zap.Any("request:", request),
		zap.Any("response", response))

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

func NewService(client *sms.Client, appId string, SignName string, l limiter.Limiter) *Service {
	return &Service{
		client:   client,
		appId:    &appId,
		signName: &SignName,
		limiter:  l,
	}

}
