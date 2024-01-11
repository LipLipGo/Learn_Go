package service

import (
	"Learn_Go/webook/internal/repository"
	"Learn_Go/webook/internal/service/sms"
	"context"
	"fmt"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeVerifyTooMany

type CodeService struct {
	repo *repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo: repo,
		sms:  smsSvc,
	}
}

func (c *CodeService) Send(ctx context.Context, biz, phone string) error {

	code := c.generate()
	err := c.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 如果验证码保存成功，则开始发送验证码
	const codeTplId = "1877556" // 可以将验证码模板设置为常量，很少改动
	err = c.sms.Send(ctx, codeTplId, []string{code}, phone)

	return err
}

func (c *CodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {

	ok, err := c.repo.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooMany {
		// 相当于，对外面屏蔽了验证次数过多的错误，只告诉调用者，验证码不对
		return false, nil
	}
	return ok, err
}

func (c *CodeService) generate() string { // 生成6位随机数验证码
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code) // 格式化随机数，保留0
}
