package ioc

import (
	"Learn_Go/webook/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitLogger() logger.LoggerV1 {
	// 使用 viper 读取 logger 的配置
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("logger", &cfg)
	if err != nil {
		panic(err)
	}
	// 使用配置创建 logger
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger.NewZapLogger(l)
}
