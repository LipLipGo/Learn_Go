package ratelimit

import (
	"Learn_Go/webook/internal/service/sms"
	"Learn_Go/webook/pkg/limiter"
	"context"
	"errors"
)

// 在这里定义一个err，一旦别人想知道这个err，就可以把这个err设置为公共的，可以提前定义好
var errLimited = errors.New("触发了限流")

type RateLimitSMSService struct {
	// 被装饰的
	svc     sms.Service
	limiter limiter.Limiter
	key     string
}

func (r *RateLimitSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limited {
		return errLimited
	}
	return r.svc.Send(ctx, tplId, args, number...)
}

func NewRateLimitSMSService(svc sms.Service, l limiter.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: l,
		key:     "sms-limiter",
	}
}
