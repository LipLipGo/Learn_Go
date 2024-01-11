package sms

import "context"

// Service 是发送短信的抽象
// 为了屏蔽不同供应商之间的区别

type Service interface {
	Send(ctx context.Context, tplId string, args []string, number ...string) error
}
