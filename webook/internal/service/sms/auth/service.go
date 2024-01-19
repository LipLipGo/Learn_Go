package auth

import (
	"Learn_Go/webook/internal/service/sms"
	"context"
	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc sms.Service
	key []byte
}

func (s *SMSService) Send(ctx context.Context, tplToken string, args []string, number ...string) error {
	var claims SMSClaims
	// 只要这里解析token通过了就行，甚至不用去校验其它东西
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	return s.svc.Send(ctx, claims.tpl, args, number...)
}

type SMSClaims struct {
	jwt.RegisteredClaims
	tpl string
}
