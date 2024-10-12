package main

import (
	"SimShare/internal/web"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	u := web.NewUserHandler()

	u.RegisterRouters(server)
	server.Run(":8080")
}
