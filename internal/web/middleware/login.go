package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		now := time.Now().Unix() //获取当前时间
		updateTime := sess.Get("updateTime")
		sess.Set("userId", id)
		sess.Options(sessions.Options{
			MaxAge: 60,
		})
		if updateTime == nil {
			// 说明是第一次登录还没有设置updateTime
			sess.Set("updateTime", now)
			sess.Save()
			return
		}
		// 非第一次登录，比较updateTime，刷新session
		updateTimeVal, _ := updateTime.(int64)

		if (now - updateTimeVal) > 10 { // 超过了10s就刷新
			sess.Set("updateTime", now)
			sess.Save()
		}
	}
}
