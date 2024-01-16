//go:build wireinject

package startup

import (
	"Learn_Go/webook/internal/repository"
	"Learn_Go/webook/internal/repository/cache"
	"Learn_Go/webook/internal/repository/dao"
	"Learn_Go/webook/internal/service"
	"Learn_Go/webook/internal/web"
	"Learn_Go/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		InitRedis, ioc.InitDB,
		// dao
		dao.NewGORMUserDao,
		// cache
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		// repository
		repository.NewCodeRepository, repository.NewCachedUserRepository,
		// service
		ioc.InitSmsService,
		service.NewuserService, service.NewcodeService,
		// handler
		web.NewUserHandler,

		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
