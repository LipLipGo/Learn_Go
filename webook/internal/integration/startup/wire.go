//go:build wireinject

package startup

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

var thirdParty = wire.NewSet(InitRedis, InitDB, InitLogger)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		thirdParty,
		// dao
		dao.NewGORMUserDao, dao.NewArticleGORMDAO,
		// cache
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		// repository
		repository.NewCodeRepository, repository.NewCachedUserRepository, repository.NewCachedArticleRepository,
		// service
		ioc.InitSmsService,
		service.NewuserService, service.NewcodeService, service.NewArticleService, InitWechatService,
		// handler
		web.NewUserHandler, web.NewArticleHandler, web.NewOAuth2WechatHandler, ijwt.NewRedisJWTHandler,

		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdParty,
		dao.NewArticleGORMDAO,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}
