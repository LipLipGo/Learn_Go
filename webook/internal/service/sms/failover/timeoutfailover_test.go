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

func TestTimeOutFailOverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name      string
		mocks     func(ctrl *gomock.Controller) []sms.Service
		threshold int32
		idx       int32 // 这里需要mock一下字段值
		cnt       int32
		wantErr   error
		wantCnt   int32 // 这里因为在调用方法时，修改了字段值，所以这里还需要判定一下字段值是否正确被修改
		wantIdx   int32
	}{
		{
			name: "没有触发切换",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0}
			},
			idx:       0,
			cnt:       12,
			threshold: 15,
			wantErr:   nil,
			// 成功了，重置了超时计数
			wantCnt: 0,
			wantIdx: 0,
		},
		{
			name: "触发切换，发送成功",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("发送失败"))

				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			idx:       0,
			cnt:       15,
			threshold: 15,
			wantErr:   nil,
			wantCnt:   0,
			wantIdx:   1,
		},
		{
			name: "触发切换，发送失败",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))

				svc1 := smsmocks.NewMockService(ctrl)
				// 如果报错missing calls，那么是因为多次调用一个方法，在报错的调用后面加上AnyTimes()即可
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("发送失败"))
				return []sms.Service{svc0, svc1}
			},
			idx:       1,
			cnt:       15,
			threshold: 15,
			wantErr:   errors.New("发送失败"),
			wantCnt:   0,
			wantIdx:   0,
		},
		{
			name: "触发切换，超时",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				return []sms.Service{svc0, svc1}
			},
			idx:       1,
			cnt:       15,
			threshold: 15,
			wantErr:   context.DeadlineExceeded,
			wantCnt:   1,
			wantIdx:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewTimeOutFailOverSMSService(tc.mocks(ctrl), tc.threshold)
			svc.idx = tc.idx // 调用方法时会修改这个字段，所以需要传入指定值
			svc.cnt = tc.cnt
			err := svc.Send(context.Background(), "123", []string{"123"}, "12345")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantIdx, svc.idx)
			assert.Equal(t, tc.wantCnt, svc.cnt)

		})
	}
}
