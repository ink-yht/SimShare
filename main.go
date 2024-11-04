package main

import (
	"SimShare/internal/repository"
	"SimShare/internal/repository/cache"
	"SimShare/internal/repository/dao"
	"SimShare/internal/service"
	"SimShare/internal/service/sms/memory"
	"SimShare/internal/web"
	middlelware "SimShare/internal/web/middleware"
	"SimShare/pkg/ginx/middleware/ratelimit"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func main() {

	//db := initDB()
	//
	//rdb := initRedis()
	//
	//server := initWebService()
	//
	//u := initUserHDL(db, rdb)
	//u.RegisterRouters(server)

	server := InitWebServer()

	err := server.Run(":8080")
	if err != nil {
		return
	}
}

func initUserHDL(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	uDAO := dao.NewUserDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(uDAO, uc)
	svc := service.NewUserService(repo)

	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)
	u := web.NewUserHandler(svc, codeSvc)
	return u
}

func initWebService() *gin.Engine {
	server := gin.Default()

	// 基于 Redis 的 IP 限流
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	server.Use(ratelimit.NewBuilder(redisClient, time.Minute, 100).Build())

	server.Use()

	// 步骤一
	//store := cookie.NewStore([]byte("secret"))
	//server.Use(sessions.Sessions("mysession", store))

	//store, err := redis.NewStore(16, "tcp", "localhost:6380", "", []byte("MKb1jQVV49L4FfjZ4QQW3rXhQ8IcaCem"))
	//if err != nil {
	//	panic(err)
	//}

	store := memstore.NewStore([]byte("MKb1jQVV49L4FfjZ4QQW3rXhQ8IcaCem"))

	server.Use(sessions.Sessions("mysession", store))

	// 步骤四
	//useSession(server)
	useJWT(server)

	return server
}

func useSession(server *gin.Engine) gin.IRoutes {
	return server.Use(middlelware.NewLoginMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/users/login").Build())
}

func useJWT(server *gin.Engine) gin.IRoutes {
	return server.Use()
}

func initDB() *gorm.DB {
	dsn := "root:root@tcp(localhost:13316)/SimShare?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	return redisClient
}
