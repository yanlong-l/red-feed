package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"redfeed/internal/web"
	"strings"
	"time"
)

func main() {
	server := gin.Default()

	// 添加跨域中间件
	server.Use(cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		//ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许你带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 你的开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	userHandler := web.NewUserHandler()
	userHandler.RegisterRoutes(server)

	err := server.Run(":8080")
	if err != nil {
		panic(err.Error())
	}
}
