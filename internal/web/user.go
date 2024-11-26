package web

import (
	"fmt"
	"net/http"
	"red-feed/internal/domain"
	"red-feed/internal/service"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/golang-jwt/jwt/v5"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

var _ Handler = &UserHandler{} // 确保userHandler实现了 handler接口

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	jwtHandler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		codeSvc:     codeSvc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)
	ug.POST("/refresh_token", u.RefreshToken)
}

func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	// 验证refresh token是否有效，如果有效，则办法一个新的access token
	// 从header中读取refresh token
	authStr := ctx.Request.Header.Get("Authorization")
	tokenSplited := strings.Split(authStr, " ")
	if authStr == "" || len(tokenSplited) != 2 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	tokenStr := tokenSplited[1]
	// 解析token
	var uc UserClaims
	token, err := jwt.ParseWithClaims(tokenStr, &uc, func(t *jwt.Token) (interface{}, error) {
		return []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf1"), nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 生成新的token
	err = u.setJWTToken(ctx, uc.Uid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "刷新成功",
	})

}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "你的邮箱格式不对")
		return
	}
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		// 记录日志
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	// 调用svc层
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "注册成功")

}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	err = u.setLoginToken(ctx, user.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	// 此时，登录成功， 我们需要维持登录态
	sess := sessions.Default(ctx)
	// 设置过期时间
	sess.Options(sessions.Options{
		MaxAge: 60,
	})
	sess.Set("userId", user.Id)
	err = sess.Save()
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Options(sessions.Options{
		//Secure: true,
		//HttpOnly: true,
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "退出登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	ctx.String(http.StatusOK, "这是你的edit")
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "这是你的profile")
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	claimsVal, _ := ctx.Get("claims")
	claims, ok := claimsVal.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 获取用户信息
	user, err := u.svc.Profile(ctx, claims.Uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	fmt.Println(user)
	ctx.String(http.StatusOK, fmt.Sprintf("用户ID:%d", claims.Uid))
}

// SendLoginSMSCode 发送验证码
func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	// 接收参数
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	err = u.codeSvc.Send(ctx, "login", req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
}

// LoginSMS 手机验证码登录
func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := u.codeSvc.Verfiy(ctx, "login", req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码不对",
		})
		return
	}
	// 验证码校验通过，FindOrCreate User
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	// 设置JWT Token
	err = u.setJWTToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})

}
