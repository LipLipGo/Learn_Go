package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	//return redis.NewClient(&redis.Options{Addr: config.Config.Redis.Addr})
	return redis.NewClient(&redis.Options{Addr: viper.GetString("redis.addr")}) // 使用 viper 读取配置
}
