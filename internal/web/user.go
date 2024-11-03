package web

import (
	"SimShare/internal/domain"
	"SimShare/internal/service"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

const biz = "login"

const (
	emailRegexPattern    = "^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$"
	passwordRegexPattern = "^(?=.*[a-zA-Z])(?=.*[0-9])(?=.*[._~!@#$^&*])[A-Za-z0-9._~!@#$^&*]{8,20}$"
)

type UserHandler struct {
	emailRexExp *regexp.Regexp
	passwordRex *regexp.Regexp
	svc         *service.UserService
	codeSvc     *service.CodeService
}

// 预编译提高校验速度

func NewUserHandler(svc *service.UserService, codeSvc *service.CodeService) *UserHandler {
	return &UserHandler{
		emailRexExp: regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRex: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:         svc,
		codeSvc:     codeSvc,
	}
}

func (h *UserHandler) RegisterRouters(server *gin.Engine) {
	UserGroup := server.Group("/users")
	UserGroup.POST("/signup", h.SignUp)
	//UserGroup.POST("/login", h.Login)
	UserGroup.POST("/login", h.LoginJWT)
	UserGroup.POST("/edit", h.Edit)
	UserGroup.GET("/profile", h.ProfileJWT)
	UserGroup.POST("/login_sms/code/send", h.SendLoginSmsCode)
	UserGroup.POST("/login_sms", h.LoginSms)
}

func (h *UserHandler) LoginSms(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 加上各种校验
	ok, err := h.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "系统错误",
			Data: nil,
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 1,
			Msg:  "验证码有误",
		})
		return
	}

	// 生成 jwt

	// 手机号会不会是一个新用户？
	user, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "系统错误",
			Data: nil,
		})
		return
	}
	if err = h.setJWT(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 2,
			Msg:  "系统错误",
			Data: nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "验证码校验通过",
		Data: nil,
	})
}

func (h *UserHandler) SendLoginSmsCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 校验是不是一个合格的手机号（正则）
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 1,
			Msg:  "输入有误",
			Data: nil,
		})
		return
	}

	err := h.codeSvc.Send(ctx, biz, req.Phone)

	if err == service.ErrCodeSetTooMany {
		ctx.JSON(http.StatusOK, Result{
			Code: 1,
			Msg:  "发送验证码太频繁",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 2,
			"msg":  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "发送成功",
		Data: nil,
	})
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

	if err == service.ErrDuplicate {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "邮箱已被注册",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 2,
			"msg":  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "注册成功",
	})
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := h.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "用户不存在或密码不对",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 2,
			"msg":  "系统错误",
		})
		return
	}

	// 步骤二
	// 登陆成功了
	// 设置 session
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Save()

	// 设置 JWT 登录态
	//claims := UserClaims{
	//	Uid: user.Id,
	//	RegisteredClaims: jwt.RegisteredClaims{
	//		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
	//	},
	//}
	//
	//token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//
	//fmt.Println(token)
	//
	//tokenStr, err := token.SignedString([]byte("hcc9ByEkfLwmRUWLFEvr2RcPXhqecE12"))
	//if err != nil {
	//	ctx.JSON(http.StatusOK, gin.H{
	//		"code": 1,
	//		"msg":  "系统异常",
	//	})
	//	return
	//}
	//
	//ctx.Header("x-jwt-token", tokenStr)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "登录成功",
	})
}
func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := h.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "用户不存在或密码不对",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 2,
			"msg":  "系统错误",
		})
		return
	}

	// 步骤二
	// 登陆成功了
	// 设置 session
	//sess := sessions.Default(ctx)
	//sess.Set("userId", user.Id)
	//sess.Save()

	if err = h.setJWT(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 2,
			"msg":  "系统异常",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "登录成功",
	})
}

func (h *UserHandler) setJWT(ctx *gin.Context, uid int64) error {
	// 设置 JWT 登录态
	claims := UserClaims{
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 55)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {

		return err
	}

	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *UserHandler) Edit(ctx *gin.Context) {

}

func (h *UserHandler) Profile(ctx *gin.Context) {

}

func (h *UserHandler) ProfileJWT(ctx *gin.Context) {
	uc := ctx.MustGet("claims").(UserClaims)
	fmt.Println(uc.Uid)
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "获取成功",
		"data": uc.Uid,
	})

	//c, _ := ctx.Get("claims")
	//
	//fmt.Println(c)
	//fmt.Println()
	//claims, ok := c.(*UserClaims)
	//if !ok {
	//	ctx.JSON(http.StatusOK, gin.H{
	//		"code": 2,
	//		"msg":  "系统错误",
	//	})
	//	return
	//}
	//fmt.Println(claims.Uid)
}

type UserClaims struct {
	jwt.RegisteredClaims
	// 声明自己要放进去 token 里的数据
	Uid       int64
	UserAgent string
}

var JWTKey = []byte("3vnkm3RPr55524y0uuG2PeEUPAT1t3PI")
