package web

import (
	"net/http"
	"red-feed/internal/service"
	"red-feed/internal/service/oauth2/wechat"

	"github.com/gin-gonic/gin"
)

var _ Handler = &OAuth2WechatHandler{} // 确保OAuth2WechatHandler实现了 handler接口

type OAuth2WechatHandler struct {
	wechatSvc wechat.Service
	userSvc   service.UserService
	jwtHandler
}

func NewOAuth2WechatHandler(wechatSvc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		wechatSvc: wechatSvc,
		userSvc:   userSvc,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

// AuthURL 构造跳转微信扫码页面的URL
func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	url, err := h.wechatSvc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造扫码登录URL失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

// Callback 处理微信扫码回调
func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	// 接收微信回调传来的code state
	code := ctx.Query("code")
	state := ctx.Query("state")
	// 校验code 获取微信用户信息
	wechatInfo, err := h.wechatSvc.VerifyCode(ctx, code, state)
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
