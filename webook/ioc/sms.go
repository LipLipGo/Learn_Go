package ioc

import (
	"Learn_Go/webook/internal/service/sms"
	"Learn_Go/webook/internal/service/sms/localsms"
	"Learn_Go/webook/internal/service/sms/tencent"
	"Learn_Go/webook/pkg/limiter"
	"github.com/redis/go-redis/v9"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"time"
)

func InitSmsService() sms.Service {
	//ratelimit.NewRateLimitSMSService(localsms.NewService(), limiter.NewRedisSlidingWindowLimiter())	// 装饰器模式
	return localsms.NewService()
	// 如果有需要，可以用这个
	//return initTencentSmsService()
}

func initTencentSmsService() sms.Service {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到腾讯 SMS 的 secret id")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("找不到腾讯 SMS 的 secret key")
	}
	c, err := tencentSMS.NewClient(
		common.NewCredential(secretId, secretKey), "ap_beijing", profile.NewClientProfile(),
	)
	if err != nil {
		panic(err)
	}
	return tencent.NewService(c, "1400842696", "Lip", limiter.NewRedisSlidingWindowLimiter(rdb, time.Second, 3000))
}
