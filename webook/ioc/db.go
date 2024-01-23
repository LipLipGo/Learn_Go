package ioc

import (
	"Learn_Go/webook/internal/repository/dao"
	"Learn_Go/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	// 在这里通过定义结构体来初始化配置项
	type Config struct {
		DSN string `yaml:"dsn"` // 这个字段对应 yaml 文件中是 dsn
	}
	var cfg = Config{
		DSN: "root:root@tcp(localhost:13316)/webook", // 建议将默认值放在这
	}

	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)

	}
	//db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			// 慢查询，设置为0，就是全部都打印
			SlowThreshold: 0,
			LogLevel:      glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

// 函数衍生类型实现接口
type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{Key: "args", Value: i})
}
