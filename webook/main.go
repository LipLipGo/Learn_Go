package main

import (
	"Learn_Go/webook/internal/web/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {

	server := InitWebServer()
	//db := initDB()
	//rd := initRedis()
	//server := initWebServer()
	//codeSvc := initCodeSvc(rd)
	//initUserHdl(db, server, rd, codeSvc)

	// 首先去除mysql和redis依赖，构造最简单的Web服务部署到k8s上

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了！")
	})

	server.Run(":8080")
}

// 下面是使用wire改造前的代码
//
//	func initUserHdl(db *gorm.DB, server *gin.Engine, rd redis.Cmdable, codeSvc *service.codeService) {
//		ud := dao.NewGORMUserDao(db)
//		uc := cache.NewRedisUserCache(rd)
//		ur := repository.NewCachedUserRepository(ud, uc)
//		us := service.NewuserService(ur)
//
//		//hdl := &web.UserHandler{}
//		hdl := web.NewUserHandler(us, codeSvc) // 设置了正则表达式预编译后需要替换为这个方法
//		hdl.RegisterRouters(server)
//	}
//
//	func initCodeSvc(redisClient redis.Cmdable) *service.codeService {
//		cc := cache.NewRedisCodeCache(redisClient)
//		cRepo := repository.NewCodeRepository(cc)
//		return service.NewcodeService(cRepo, initMemorySms())
//
// }
//
//	func initMemorySms() sms.Service {
//		return localsms.NewService()
//	}
//
//	func initWebServer() *gin.Engine {
//		server := gin.Default()
//
//		// 解决跨域问题
//		// 跨域请求：请求是从 localhost:3000 前端发送到 localhost:8080 后端的
//		// 类似这种就是跨域请求，协议、域名和端口任意一个不同，都是跨域请求
//		// 如果不做额外处理，没办法发送请求的	；浏览器会发送一个预检请求 preflight 给后端，询问是否接收请求
//
//		/*跨域问题要点：
//		1.跨域问题是因为发请求的 协议+域名+端口 和接受请求的 协议+域名+端口 对不上，比如 localhost:3000 发到 localhost:8080  上
//		2.解决跨域问题的关键是在 preflight 请求里告诉浏览器自己愿意接受请求
//		3.Gin 提供了解决跨域问题的 middleware ，可以直接使用
//		4.middleware 是一种机制，可以用来解决一些所有业务都关心的问题，使用 Use 方法来注册 middleware
//		*/
//
//		server.Use(cors.New(cors.Config{ // 通过 Middleware（cors） 处理跨域请求
//			//AllowAllOrigins: true,	允许所有的源头
//			//AllowOrigins: []string{"http://localhost:3000"}, //允许一些
//			//AllowMethods: []string{"POST"},   最好不要设置，允许所有请求方法即可
//			AllowCredentials: true,                                      // cookie的数据是否允许传过来，正常情况下允许
//			AllowHeaders:     []string{"Content-Type", "Authorization"}, //报错，根据报错找到需要添加的headers
//			// 允许前端访问后端响应中带的头部
//			ExposeHeaders: []string{"x-jwt-token"},
//			AllowOriginFunc: func(origin string) bool {
//				if strings.HasPrefix(origin, "http://localhost") {
//					return true
//				}
//
//				return strings.Contains(origin, "your_company.com")
//			},
//			MaxAge: 12 * time.Hour, // preflight检测时长，无影响
//		}), func(ctx *gin.Context) {
//			fmt.Println("这是一个 Middleware")
//		})
//
//		//redisClient := redis.NewClient(&redis.Options{
//		//	Addr: config.Config.Redis.Addr,
//		//})
//		//// 一秒钟100次
//		//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
//
//		//useSession(server)
//		useJWT(server)
//
//		return server
//	}
//
//	func useJWT(server *gin.Engine) {
//		loginJWT := middleware.LoginJWTMiddlewareBuilder{}
//		server.Use(loginJWT.CheckLogin())
//	}
func useSession(server *gin.Engine) {
	// 登陆校验
	login := &middleware.LoginMiddlewareBuilder{}
	// 存储数据的，也就是 userId 存哪里
	// 1.直接存 Cookie
	store := cookie.NewStore([]byte("secret"))

	// 2.使用 memstore 内存的实现， 需要传入两个key ，第一个用于 authentication 用于身份认证，第二个用于 encryption 用于数据加密，可通过工具生成
	// store := memstore.NewStore([]byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7lJ"), []byte("yY3aL4pO2iQ5vA5jE9yQ0vN1sC2vW4rN"))

	// 3.使用 redis 实现，需要启动 redis 服务，传入参数 size 表示连接数 + network “tcp” + address “主机+端口” + password 未设置 + 两个key
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	//	[]byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7lJ"), []byte("yY3aL4pO2iQ5vA5jE9yQ0vN1sC2vW4rN"))
	//if err != nil {
	//	panic(err)
	//}
	server.Use(sessions.Sessions("ssid", store), login.CheckLogin()) // 先初始化session，再校验

}

//后端处理  1.接受请求并校验	 2.调用业务逻辑处理请求  3.根据业务逻辑处理结果返回响应
