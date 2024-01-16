// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"Learn_Go/webook/internal/repository"
	"Learn_Go/webook/internal/repository/cache"
	"Learn_Go/webook/internal/repository/dao"
	"Learn_Go/webook/internal/service"
	"Learn_Go/webook/internal/web"
	"Learn_Go/webook/ioc"
	"github.com/gin-gonic/gin"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	v := ioc.InitGinMiddleWares(cmdable)
	db := ioc.InitDB()
	userDao := dao.NewGORMUserDao(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDao, userCache)
	userService := service.NewuserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSmsService()
	codeService := service.NewcodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService)
	engine := ioc.InitWebServer(v, userHandler)
	return engine
}
