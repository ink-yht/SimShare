package main

import (
	"SimShare/internal/repository"
	"SimShare/internal/repository/dao"
	"SimShare/internal/service"
	"SimShare/internal/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {

	db := initDB()

	server := initWebService()

	initUserHDL(db, server)

	err := server.Run(":8080")
	if err != nil {
		return
	}
}

func initUserHDL(db *gorm.DB, server *gin.Engine) {
	uDAO := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(uDAO)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	u.RegisterRouters(server)
}

func initWebService() *gin.Engine {
	server := gin.Default()
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
	return server
}

func initDB() *gorm.DB {
	dsn := "root:root@tcp(127.0.0.1:3306)/SimShare?charset=utf8mb4&parseTime=True&loc=Local"
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
