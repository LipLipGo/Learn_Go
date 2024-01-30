package startup

import (
	"Learn_Go/webook/internal/service/oauth2/wechat"
	"Learn_Go/webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {

	return wechat.NewWechatService("", "", l)
}
