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
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {

	db := initDB()

	rdb := initRedis()

	server := initWebService()

	initUserHDL(db, rdb, server)

	err := server.Run(":8080")
	if err != nil {
		return
	}
}

func initUserHDL(db *gorm.DB, rdb redis.Cmdable, server *gin.Engine) {
	uDAO := dao.NewUserDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(uDAO, uc)
	svc := service.NewUserService(repo)

	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)
	u := web.NewUserHandler(svc, codeSvc)
	u.RegisterRouters(server)
}

func initWebService() *gin.Engine {
	server := gin.Default()

	// 基于 Redis 的 IP 限流
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	server.Use(ratelimit.NewBuilder(redisClient, time.Minute, 100).Build())

	server.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"https://foo.com"},
		//AllowMethods:     []string{"PUT", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your.com")
		},
		MaxAge: 12 * time.Hour,
	}))

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
	return server.Use(middlelware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/users/login").Build())
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
