package ratelimit

import (
	"Learn_Go/webook/internal/service/sms"
	smsmocks "Learn_Go/webook/internal/service/sms/mocks"
	"Learn_Go/webook/pkg/limiter"
	limitermocks "Learn_Go/webook/pkg/limiter/mocks"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRateLimitSMSService_Send(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter)

		// 这里测试限流，与这些输入没啥关系，所以在这里可以不定义输入，在运行测试用例的时候写死
		//ctx    context.Context
		//tplId  string
		//args   []string
		//number []string

		// 但是预期输出还是要有
		wantErr error
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return svc, l
			},
			wantErr: nil,
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return svc, l
			},
			wantErr: errLimited,
		},
		{
			name: "限流器错误",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, errors.New("限流器错误"))
				return svc, l
			},
			wantErr: errors.New("限流器错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			smsSvc, l := tc.mock(ctrl)
			svc := NewRateLimitSMSService(smsSvc, l)
			err := svc.Send(context.Background(), "abc", []string{"123"}, "123")

			assert.Equal(t, tc.wantErr, err)
		})
	}
}
