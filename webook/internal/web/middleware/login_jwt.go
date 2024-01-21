package middleware

import (
	ijwt "Learn_Go/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type LoginJWTMiddlewareBuilder struct {
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(hdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: hdl,
	}
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			// 不需要登录校验
			return
		}

		tokenStr := m.ExtractToken(ctx)
		var uc ijwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) { // 这里 uc 需要传指针
			return ijwt.JWTKey, nil
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

		// 在这里去校验 ssid 是否失效，因为我们可以在前面先校验一下 token ，避免无效的查询 Redis

		//cnt, err := m.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid)).Result()
		err = m.CheckSession(ctx, uc.Ssid) // jwt 改造成面向接口编程之后，这里面判断了 ssid 的有效性

		// 这种判定方式过于严格，因为一旦 redis 崩溃了，就无法继访问服务
		if err != nil {
			// ssid 无效或者 redis 有问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 这种写法可以兼容 redis 异常的情况，就是即便 redis 崩溃了，但是用户依然可以访问服务
		// 但是要做好监控，有没有 error
		//if cnt > 0 {
		//	// ssid 无效
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		// *****************************************************************************************
		// 这里是自动刷新JWTtoken，但是我们使用了长短token机制，所以自动刷新机制需要屏蔽
		//// 刷新登陆状态
		//// 获取刷新时间
		//expireTime := uc.ExpiresAt
		//// 这里判断是否过期其实前面 Valid 就能够实现，这里不判定也可以
		////if expireTime.Before(time.Now()) {
		////	ctx.AbortWithStatus(http.StatusUnauthorized)
		////	return
		////}
		//
		//if expireTime.Sub(time.Now()) < time.Second*50 {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30)) // 记录新的过期时间
		//	newToken, err := token.SignedString(web.JWTKey)                     // 获取新的token
		//	ctx.Header("x-jwt-token", newToken)                                 // 传入新的 token
		//	if err != nil {
		//		log.Println(err) // 如果刷新没成功，不影响登陆状态，不用中断
		//	}
		//
		//}
		// *******************************************************************************************
		ctx.Set("user", uc) // 设置缓存，节省时间，后续可直接获取uc

	}
}
