package main

import (
	"red-feed/config"
	"red-feed/internal/repository"
	"red-feed/internal/repository/cache"
	"red-feed/internal/repository/dao"
	"red-feed/internal/service"
	"red-feed/internal/service/sms/memory"
	"red-feed/internal/web"
	"red-feed/internal/web/middleware"
	"red-feed/pkg/ginx/middlewares/ratelimit"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	rdb := initRedis()
	server := initWebServer(rdb) // 初始化web server

	u := initUser(db, rdb)
	u.RegisterRoutes(server)

	server.Run(":8080")
}

func initWebServer(rdb redis.Cmdable) *gin.Engine {
	server := gin.Default()

	server.Use(ratelimit.NewBuilder(rdb, time.Second, 200).Build())
	server.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"x-jwt-token"},
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

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/login").
		IgnorePaths("/users/login").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").Build())

	return server
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, uc)
	svc := service.NewUserService(repo)
	smsSvc := memory.NewService()
	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	codeSvc := service.NewCodeService(smsSvc, codeRepo)
	u := web.NewUserHandler(svc, codeSvc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		// 我只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		// 一旦初始化过程出错，应用就不要启动了
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
}
