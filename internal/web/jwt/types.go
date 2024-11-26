package ijwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	SetJWTToken(ctx *gin.Context, uid int64, ssid string) error
	SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error
	SetLoginToken(ctx *gin.Context, uid int64) error
	ClearToken(ctx *gin.Context) error
	CheckSession(ctx *gin.Context, ssid string) error
	ExtractToken(ctx *gin.Context) string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string // session_id 标识一次登录回话是否有效
	UserAgent string
}
