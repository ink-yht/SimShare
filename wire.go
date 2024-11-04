//go:build wireinject

package main

import (
	"SimShare/internal/repository"
	"SimShare/internal/repository/cache"
	"SimShare/internal/repository/dao"
	"SimShare/internal/service"
	"SimShare/internal/web"
	"SimShare/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,

		// DAO 部分
		dao.NewUserDAO,

		// cache 部分
		cache.NewUserCache,
		cache.NewCodeCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,

		// service 部分
		service.NewUserService,
		service.NewCodeService,
		// 直接基于内存实现 （手机号注册）
		ioc.InitSmsService,

		// Handler 部分
		web.NewUserHandler,

		// 中间件
		ioc.InitWebServer,
		ioc.InitMiddleWares,
	)
	return new(gin.Engine)
}
