package ioc

import (
	"Learn_Go/webook/internal/web"
	ijwt "Learn_Go/webook/internal/web/jwt"
	"Learn_Go/webook/internal/web/middleware"
	"Learn_Go/webook/pkg/ginx/middleware/ratelimit"
	"Learn_Go/webook/pkg/limiter"
	"Learn_Go/webook/pkg/logger"
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, authHdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRouters(server)
	authHdl.RegisterRoutes(server)
	articleHdl.RegisterRouters(server)
	return server

}

func InitGinMiddleWares(redisClient redis.Cmdable, hdl ijwt.Handler, l logger.LoggerV1) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{ // 通过 Middleware（cors） 处理跨域请求
			//AllowAllOrigins: true,	允许所有的源头
			//AllowOrigins: []string{"http://localhost:3000"}, //允许一些
			//AllowMethods: []string{"POST"},   最好不要设置，允许所有请求方法即可
			AllowCredentials: true,                                      // cookie的数据是否允许传过来，正常情况下允许
			AllowHeaders:     []string{"Content-Type", "Authorization"}, //报错，根据报错找到需要添加的headers
			// 允许前端访问后端响应中带的头部
			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					return true
				}

				return strings.Contains(origin, "your_company.com")
			},
			MaxAge: 12 * time.Hour, // preflight检测时长，无影响
		}), func(ctx *gin.Context) {
			fmt.Println("这是一个 Middleware")
		},
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 100)).Build(),
		middleware.NewLogMiddlewareBuilder(func(ctx context.Context, lc middleware.LogContent) {
			l.Debug("", logger.Field{Key: "req", Value: lc})
		}).AllowReqBody().AllowRespBody().Build(),
		middleware.NewLoginJWTMiddlewareBuilder(hdl).CheckLogin(),

		// 使用 session 登录校验

		//sessions.Sessions("ssid", cookie.NewStore([]byte("secret"))),
		//(&middleware.LoginMiddlewareBuilder{}).CheckLogin(),
	}
}
