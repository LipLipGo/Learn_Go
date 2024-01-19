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
		dao.NewGORMUserDao,
		// cache
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		// repository
		repository.NewCodeRepository, repository.NewCachedUserRepository,
		// service
		ioc.InitSmsService,
		ioc.InitWechatService,
		service.NewuserService, service.NewcodeService,
		// handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,

		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
