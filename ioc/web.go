package ioc

import (
	"context"
	"red-feed/internal/web"
	ijwt "red-feed/internal/web/jwt"
	"red-feed/internal/web/middleware"
	"red-feed/pkg/ginx/middlewares/logger"
	"red-feed/pkg/ginx/middlewares/metric"
	"red-feed/pkg/ginx/middlewares/ratelimit"
	ilogger "red-feed/pkg/logger"
	pkg_ratelimit "red-feed/pkg/ratelimit"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler,
	artHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler, l ilogger.Logger) []gin.HandlerFunc {
	limiter := pkg_ratelimit.NewRedisSlidingWindowLimiter(redisClient, 200, time.Second)
	return []gin.HandlerFunc{
		ratelimit.NewBuilder(limiter).Build(),
		//ginx.InitCounter(prometheus.CounterOpts{
		//	Namespace: "internal_test",
		//	Subsystem: "red_feed",
		//	Name:      "http_biz_code",
		//	Help:      "GIN 中 HTTP 请求",
		//	ConstLabels: map[string]string{
		//		"instance_id": "my-instance-1",
		//	},
		//}),
		otelgin.Middleware("red_feed"),
		(&metric.MiddlewareBuilder{
			Namespace:  "internal_test",
			Subsystem:  "red_feed",
			Name:       "gin_http",
			Help:       "统计 GIN 的 HTTP 接口",
			InstanceID: "my-instance-1",
		}).Build(),
		logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
			l.Info("access log", ilogger.Field{Key: "access log desc", Value: al})
		}).AllowReqBody(true).
			AllowRespBody(true).Build(),
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
