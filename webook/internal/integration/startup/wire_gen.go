// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"Learn_Go/webook/internal/repository"
	"Learn_Go/webook/internal/repository/cache"
	"Learn_Go/webook/internal/repository/dao"
	"Learn_Go/webook/internal/service"
	"Learn_Go/webook/internal/web"
	"Learn_Go/webook/internal/web/jwt"
	"Learn_Go/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := InitLogger()
	v := ioc.InitGinMiddleWares(cmdable, handler, loggerV1)
	db := InitDB()
	userDao := dao.NewGORMUserDao(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDao, userCache)
	userService := service.NewuserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSmsService()
	codeService := service.NewcodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	wechatService := InitWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	articleDAO := dao.NewArticleGORMDAO(db)
	articleRepository := repository.NewCachedArticleRepository(articleDAO)
	articleService := service.NewArticleService(articleRepository)
	articleHandler := web.NewArticleHandler(articleService, loggerV1)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

func InitArticleHandler() *web.ArticleHandler {
	db := InitDB()
	articleDAO := dao.NewArticleGORMDAO(db)
	articleRepository := repository.NewCachedArticleRepository(articleDAO)
	articleService := service.NewArticleService(articleRepository)
	loggerV1 := InitLogger()
	articleHandler := web.NewArticleHandler(articleService, loggerV1)
	return articleHandler
}

// wire.go:

var thirdParty = wire.NewSet(InitRedis, InitDB, InitLogger)
