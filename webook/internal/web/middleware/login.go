package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc { // Builder模式
	return func(ctx *gin.Context) {

		gob.Register(time.Now())

		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}

		sess := sessions.Default(ctx)
		userId := sess.Get("userId") // 获取userId，否则后面刷新登陆状态时，会将session数据覆盖
		if userId == nil {
			// 查找sess_id，如果没有，就中断，不要往后执行，也就是不执行后面的业务逻辑
			ctx.AbortWithStatus(http.StatusUnauthorized) // http.StatusUnauthorized 通常用于代表没登陆
			return
		}

		now := time.Now() // 获取当前时间戳
		// 怎么知道需要刷新了？
		// 假如一分钟刷新一次，怎么知道已经过了一分钟
		const updateTimeKey = "update_time"

		val := sess.Get(updateTimeKey)

		lastUpdateTime, ok := val.(time.Time) // 断言获取上一次刷新时间，若断言失败，重新设置

		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Second*10 {
			// 如果为空，则是第一次进来
			sess.Set(updateTimeKey, now) //  报错 gob: type not registered for interface: time.Time，redis存储数据是将结构体序列化为字节切片进行存储，需通过gob注册一下time.Now()
			sess.Set("userId", userId)
			err := sess.Save()
			if err != nil {
				fmt.Println(err) //这里不终止运行，而是打日志，因为刷新登陆状态失败并不影响正常业务进行
			}
		}

	}
}

// 然后就接入middleware
