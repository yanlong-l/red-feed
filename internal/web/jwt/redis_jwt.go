package ijwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	AtKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")
	RtKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvfx")
)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{cmd: cmd}
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	// 登录成功，生成JWT Token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	// 登录成功，生成JWT Token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	// 生成一个session id,标识这一次登录成功的会话
	ssid := uuid.New().String()
	err := h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return h.SetRefreshToken(ctx, uid, ssid)
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	// 从ctx中读出ssid，并把ssid存到redis中来标识对应的token已经不可用
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	claims := ctx.MustGet("claims")
	uc, _ := claims.(UserClaims)
	return h.cmd.Set(ctx.Request.Context(), fmt.Sprintf("users:ssid:%s", uc.Ssid), "", time.Hour*24*7).Err()
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	res, err := h.cmd.Exists(ctx.Request.Context(), fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch err {
	case redis.Nil:
		return nil
	case nil:
		if res == 0 {
			return nil
		}
		return errors.New("session已过期")
	default:
		return err
	}
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	// 从header中读取refresh token
	authStr := ctx.Request.Header.Get("Authorization")
	tokenSplited := strings.Split(authStr, " ")
	if authStr == "" || len(tokenSplited) != 2 {
		return ""
	}
	tokenStr := tokenSplited[1]
	return tokenStr
}
