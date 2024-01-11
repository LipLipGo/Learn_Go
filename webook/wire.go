//go:build wireinject

package main

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
		ioc.InitRedis, ioc.InitDB,
		// dao
		dao.NewUserDao,
		// cache
		cache.NewUserCache, cache.NewCodeCache,
		// repository
		repository.NewCodeRepository, repository.NewUserRepository,
		// service
		ioc.InitSmsService,
		service.NewUserService, service.NewCodeService,
		// handler
		web.NewUserHandler,

		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
