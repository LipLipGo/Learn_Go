package failover

import (
	"Learn_Go/webook/internal/service/sms"
	"context"
	"errors"
	"log"
	"sync/atomic"
)

type FailOverSMSService struct {
	svcs []sms.Service

	// 第二种实现字段
	// 当前服务商下标
	idx uint64
}

func NewFailOverSMSService(smsSvcs []sms.Service) *FailOverSMSService {
	return &FailOverSMSService{
		svcs: smsSvcs,
	}
}

// 每次都从第1个开始轮询

func (f *FailOverSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	// 轮询
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, number...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}

// 起始下标轮询
// 并且出错也轮询

func (f *FailOverSMSService) SendV1(ctx context.Context, tplId string, args []string, number ...string) error {
	// 这里的 idx 变化需要考虑线程安全问题
	//idx := f.idx
	idx := atomic.AddUint64(&f.idx, 1) // 原子操作，都是操作指针
	length := uint64(len(f.svcs))
	// 迭代 length 次
	for i := idx; i < idx+length; i++ {
		// 使用取余作为下标
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tplId, args, number...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:
			// 前者是主动取消发送，后者是超时
			return err
		}
		log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}
