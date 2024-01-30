package startup

import "Learn_Go/webook/pkg/logger"

func InitLogger() logger.LoggerV1 {
	return logger.NewNopLogger()
}
