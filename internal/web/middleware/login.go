package middlelware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 步骤三
		// 不需要登录的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		//if ctx.Request.URL.Path == "/users/login" || ctx.Request.URL.Path == "/users/signup" {
		//	return
		//}
		sess := sessions.Default(ctx)
		if sess == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//now := time.Now()
		//
		//// 进行判断，如何知道已经过了多长时间
		//const UpdateTimeKey = "update_time"
		//// 获取上次更新时间
		//// 如果没有，就设置一个
		//// 如果有，就判断是否超过10分钟
		//// 如果超过10分钟，就更新
		//// 如果没有超过10分钟，就不用更新
		//val := sess.Get(UpdateTimeKey)
		//lastUpdateTime, ok := val.(time.Time)
		//
		//if val == nil || (!ok) || now.Sub(lastUpdateTime) > time.Second*10 {
		//	// 第一次进来
		//	sess.Set(UpdateTimeKey, now)
		//	sess.Set("userId", id)
		//	err := sess.Save()
		//	if err != nil {
		//		// 打日志
		//		fmt.Println(err)
		//	}
		//}
	}
}
