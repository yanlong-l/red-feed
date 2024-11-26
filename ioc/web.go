package ioc

import (
	"red-feed/internal/web"
	ijwt "red-feed/internal/web/jwt"
	"red-feed/internal/web/middleware"
	"red-feed/pkg/ginx/middlewares/ratelimit"
	pkg_ratelimit "red-feed/pkg/ratelimit"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, oauth2WechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	limiter := pkg_ratelimit.NewRedisSlidingWindowLimiter(redisClient, 200, time.Second)
	return []gin.HandlerFunc{
		ratelimit.NewBuilder(limiter).Build(),
		corsHandlerFunc(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/login").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/users/signup").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").Build(),
	}
}

func corsHandlerFunc() gin.HandlerFunc {
	return cors.New(cors.Config{
		// AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
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
	})
}
