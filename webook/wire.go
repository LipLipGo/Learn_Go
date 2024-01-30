//go:build wireinject

package main

import (
	"Learn_Go/webook/internal/repository"
	"Learn_Go/webook/internal/repository/cache"
	"Learn_Go/webook/internal/repository/dao"
	"Learn_Go/webook/internal/service"
	"Learn_Go/webook/internal/web"
	ijwt "Learn_Go/webook/internal/web/jwt"
	"Learn_Go/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitLogger,
		// dao
		dao.NewGORMUserDao,
		dao.NewArticleGORMDAO,
		// cache
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		// repository
		repository.NewCodeRepository, repository.NewCachedUserRepository, repository.NewCachedArticleRepository,
		// service
		ioc.InitSmsService,
		ioc.InitWechatService,
		service.NewuserService, service.NewcodeService, service.NewArticleService,

		// handler
		ijwt.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
