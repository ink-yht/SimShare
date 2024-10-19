package middlelware

import (
	"SimShare/internal/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 步骤三
		// 不需要登录的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// jwt
		tokenHeader := ctx.Request.Header.Get("Authorization")
		if tokenHeader == "" {
			// 没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		seg := strings.SplitN(tokenHeader, " ", 2)
		if len(seg) != 2 || seg[0] != "Bearer" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := seg[1]

		fmt.Println(tokenHeader)

		claims := &web.UserClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("hcc9ByEkfLwmRUWLFEvr2RcPXhqecE12"), nil
		})
		if err != nil {

			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token == nil || token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
