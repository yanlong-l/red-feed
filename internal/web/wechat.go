package web

import (
	"errors"
	"fmt"
	"net/http"
	"red-feed/internal/service"
	"red-feed/internal/service/oauth2/wechat"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var _ Handler = &OAuth2WechatHandler{} // 确保OAuth2WechatHandler实现了 handler接口

type OAuth2WechatHandler struct {
	wechatSvc wechat.Service
	userSvc   service.UserService
	stateKey  []byte
	jwtHandler
}

func NewOAuth2WechatHandler(wechatSvc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		wechatSvc: wechatSvc,
		userSvc:   userSvc,
		stateKey:  []byte("95osj3fUD7foxmlYdDbncXz4VD2igvf1"),
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	// 生成携带state信息的token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, StateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
		State: state,
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	// 设置cookie
	ctx.SetCookie("jwt-state", tokenStr,
		600, "/oauth2/wechat/callback",
		"", true, true)

	return nil
}

// AuthURL 构造跳转微信扫码页面的URL
func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	// 生成一个state，因为我们一会要存到cookie
	state := uuid.New().String()
	url, err := h.wechatSvc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造扫码登录URL失败",
		})
		return
	}
	// 存入cookie，使得微信回调的时候，能拿到这个state
	err = h.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) VerifyState(ctx *gin.Context, state string) error {
	// 从cookie拿到state token
	tokenStr, err := ctx.Cookie("jwt-state")
	if err != nil {
		return err
	}
	// 从token中解码state 比较是否一致
	var sc StateClaims
	token, err := jwt.ParseWithClaims(tokenStr, &sc, func(t *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("state token已过期 %w", err)
	}
	if sc.State != state {
		return errors.New("state不匹配")
	}
	return nil
}

// Callback 处理微信扫码回调
func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	// 接收微信回调传来的code state
	code := ctx.Query("code")
	state := ctx.Query("state")
	// 校验state
	err := h.VerifyState(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "登录失败",
		})
		return
	}

	// 校验code 获取微信用户信息
	wechatInfo, err := h.wechatSvc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 查找或创建用户拿到Uid
	user, err := h.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 生成jwt token并设置header
	err = h.setJWTToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string `json:"state"`
}
