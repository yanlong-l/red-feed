package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler() *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/login", u.Login)
	ug.POST("/signup", u.SignUp)
	ug.POST("/edit", u.Edit)
	ug.POST("/profile", u.Profile)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	// 接收参数
	type SignUpReq struct {
		Email           string `json:"email" required:"true"`
		Password        string `json:"password" required:"true"`
		ConfirmPassword string `json:"confirmPassword" required:"true"`
	}

	var req SignUpReq

	// 参数接收
	err := ctx.Bind(&req) // 注意这里bind和should bind的区别
	if err != nil {
		return
	}

	// 参数校验
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

	// 数据库操作
	ctx.String(http.StatusOK, "Ok")
}

func (u *UserHandler) Login(ctx *gin.Context) {

}

func (u *UserHandler) Profile(ctx *gin.Context) {

}

func (u *UserHandler) Edit(ctx *gin.Context) {

}
