package web

import (
	"SimShare/internal/domain"
	"SimShare/internal/service"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	emailRegexPattern    = "^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$"
	passwordRegexPattern = "^(?=.*[a-zA-Z])(?=.*[0-9])(?=.*[._~!@#$^&*])[A-Za-z0-9._~!@#$^&*]{8,20}$"
)

type UserHandler struct {
	emailRexExp *regexp.Regexp
	passwordRex *regexp.Regexp
	svc         *service.UserService
}

// 预编译提高校验速度

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRexExp: regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRex: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:         svc,
	}
}

func (h *UserHandler) RegisterRouters(server *gin.Engine) {
	UserGroup := server.Group("/users")
	UserGroup.POST("/signup", h.SignUp)
	UserGroup.POST("/login", h.Login)
	UserGroup.POST("/edit", h.Edit)
	UserGroup.POST("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	fmt.Println(req)

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 2,
			"msg":  "系统错误",
		})
		return
	}

	if !isEmail {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "非法邮箱格式",
		})
		return
	}

	if req.ConfirmPassword != req.Password {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "两次输入密码不一致",
		})
		return
	}

	isPassword, err := h.passwordRex.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 2,
			"msg":  "系统错误",
		})
		return
	}

	if !isPassword {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "密码必须包含数字、字母、特殊字符，且不少于八位",
		})
		return
	}

	err = h.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	// err 有两种情况
	// 1.系统错误
	// 2.邮箱已注册

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 2,
			"msg":  "系统错误",
		})
		return
	}

	if err == service.ErrDuplicateEmail {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "邮箱已被注册",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "登录成功",
	})
}

func (h *UserHandler) Login(ctx *gin.Context) {

}

func (h *UserHandler) Edit(ctx *gin.Context) {

}

func (h *UserHandler) Profile(ctx *gin.Context) {

}
