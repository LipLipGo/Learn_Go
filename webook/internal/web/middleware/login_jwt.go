package middleware

import (
	"Learn_Go/webook/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}

		// 根据约定，token 在Authorization 头部
		// Bearer XXXXXX
		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized) // 如果没有这个头部token，就是没有登陆
			return
		}
		segs := strings.Split(authCode, " ") // 按照空格切割字符串，头部分为两部分 Bearer XXXXXX

		if len(segs) != 2 { // 如果不是两部分，则不对
			// 没登陆，Authorization 中的内容是乱传的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) { // 这里 uc 需要传指针
			return web.JWTKey, nil
		})
		if err != nil {
			// token不对，token是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token == nil || !token.Valid { // 这里其实 Valid 校验就可以了，包括过期
			// token解析出来了，但是token可能是非法的，或者是过期了
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if uc.UserAgent != ctx.GetHeader("User-Agent") {

			// 后期讲到监控告警的时候，这个地方要埋点
			// 能够进来这个分支的，大概率是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 刷新登陆状态
		// 获取刷新时间
		expireTime := uc.ExpiresAt
		// 这里判断是否过期其实前面 Valid 就能够实现，这里不判定也可以
		//if expireTime.Before(time.Now()) {
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		if expireTime.Sub(time.Now()) < time.Second*50 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute)) // 记录新的过期时间
			newToken, err := token.SignedString(web.JWTKey)                // 获取新的token
			ctx.Header("x-jwt-token", newToken)                            // 传入新的 token
			if err != nil {
				log.Println(err) // 如果刷新没成功，不影响登陆状态，不用中断
			}

		}

		ctx.Set("user", uc) // 设置缓存，节省时间，后续可直接获取uc

	}
}
