package ioc

import (
	"Learn_Go/webook/internal/service/oauth2/wechat"
	"os"
)

func InitWechatService() wechat.Service {
	//appId, ok := os.LookupEnv("WECHAT_APP_ID")
	//if !ok {
	//	panic("未找到 WECHAT_APP_ID")
	//}
	//appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	//if !ok {
	//	panic("未找到 WECHAT_APP_SECRET")
	//}
	appId := os.Getenv("WECHAT_APP_ID")
	if appId == "" {
		panic("找不到 WECHAT_APP_ID")
	}
	appSecret := os.Getenv("WECHAT_APP_SECRET")
	if appSecret == "" {
		panic("找不到 WECHAT_APP_SECRET")
	}
	return wechat.NewWechatService(appId, appSecret)
}
