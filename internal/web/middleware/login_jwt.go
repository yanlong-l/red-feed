package middleware

import (
	"net/http"
	ijwt "red-feed/internal/web/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: jwtHdl,
	}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		// 校验是否带有jwt token, 从请求头的authorization中解析
		tokenStr := l.ExtractToken(ctx)
		uc := &ijwt.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.AtKey, nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 校验token是否有效
		if !token.Valid || token == nil || uc.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 校验Ssid是否已经被标识为过期了
		err = l.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// 要么 redis 有问题，要么已经退出登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if ctx.Request.UserAgent() != uc.UserAgent {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 有了长短token机制，在这里取消token自动刷新机制
		// 每10s刷新一次token
		// now := time.Now()
		// expireTime := uc.RegisteredClaims.ExpiresAt.Time
		// if expireTime.Sub(now) < time.Second*50 { // 已经超过10s了
		// 	claims := web.UserClaims{
		// 		RegisteredClaims: jwt.RegisteredClaims{
		// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		// 		},
		// 		Uid:       uc.Uid,
		// 		UserAgent: ctx.Request.UserAgent(),
		// 	}
		// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		// 	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
		// 	if err != nil {
		// 		ctx.AbortWithStatus(http.StatusInternalServerError)
		// 		return
		// 	}
		// 	ctx.Header("x-jwt-token", tokenStr)
		// }
		ctx.Set("claims", uc)
	}
}
