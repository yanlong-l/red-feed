package main

import (
	"github.com/gin-gonic/gin"
	"redfeed/internal/web"
)

func main() {
	server := gin.Default()

	userHandler := web.NewUserHandler()
	userHandler.RegisterRoutes(server)

	err := server.Run(":8080")
	if err != nil {
		panic(err.Error())
	}
}
