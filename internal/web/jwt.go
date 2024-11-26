package web

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type jwtHandler struct {
}

func (j *jwtHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	// 登录成功，生成JWT Token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}


func (j *jwtHandler) setRefreshToken(ctx *gin.Context, uid int64) error {
	// 登录成功，生成JWT Token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf1"))
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}


func (j *jwtHandler) setLoginToken(ctx *gin.Context, uid int64) error {
	err := j.setJWTToken(ctx, uid); if err != nil {
		return err
	}
	return j.setRefreshToken(ctx, uid)
}


type UserClaims struct {
	jwt.RegisteredClaims
	// 声明你自己的要放进去 token 里面的数据
	Uid int64
	// 自己随便加
	UserAgent string
}
