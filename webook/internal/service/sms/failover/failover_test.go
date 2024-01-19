package failover

import (
	"Learn_Go/webook/internal/service/sms"
	smsmocks "Learn_Go/webook/internal/service/sms/mocks"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFailOverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name  string
		mocks func(ctrl *gomock.Controller) []sms.Service

		// 这里测试也可以将输入写死，与这些输入关系都不大

		wantErr error
	}{
		{
			name: "一次发送成功",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				return []sms.Service{svc0}
			},
			wantErr: nil,
		},
		{
			name: "第二次发送成功",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			wantErr: nil,
		},
		{
			name: "全部发送失败",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))
				svc2 := smsmocks.NewMockService(ctrl)
				svc2.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("轮询了所有的服务商，但是发送都失败了"))

				return []sms.Service{svc0, svc1, svc2}
			},
			wantErr: errors.New("轮询了所有的服务商，但是发送都失败了"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewFailOverSMSService(tc.mocks(ctrl))
			err := svc.Send(context.Background(), "123", []string{"123", "1233"})

			assert.Equal(t, tc.wantErr, err)

		})
	}
}