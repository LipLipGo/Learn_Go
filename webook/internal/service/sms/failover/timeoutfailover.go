package failover

import (
	"Learn_Go/webook/internal/service/sms"
	"context"
	"sync/atomic"
)

type TimeOutFailOverSMSService struct {
	svcs []sms.Service
	// 当前服务商下标
	idx int32
	// 连续几个超时
	cnt int32
	// 切换的阈值，只读，不用考虑线程安全的问题
	threshold int32
}

func NewTimeOutFailOverSMSService(svcs []sms.Service, threshold int32) *TimeOutFailOverSMSService {
	return &TimeOutFailOverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}

func (t *TimeOutFailOverSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	// 超出阈值，执行切换
	if cnt >= t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))             // 这是切换后的服务商下标
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) { // 多个用户进来这里，但是只有一个用户能够切换成功，并且将新的下标写回idx
			// 切换成功后，重置计数
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newIdx
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, number...)
	switch err {
	case nil:
		// 如果没有返回错误，那么就是发送成功，那么计数重置0
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		// 如果是超时错误，那么计数加1
		atomic.AddInt32(&t.cnt, 1)
	default:
		// 这里返回了错误，但是不是超时错误，要考虑怎么弄
		// 可以考虑增加atomic计数
		// 如果强调一定是超时，那么就不增加
		// 如果是EOF之类的错误，可以考虑直接切换

	}
	return err
}
